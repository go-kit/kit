package awssqs_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/transport/awssqs"
)

type testReq struct {
	Squadron int `json:"s"`
}

type testRes struct {
	Squadron int    `json:"s"`
	Name     string `json:"n"`
}

var names = map[int]string{
	424: "tiger",
	426: "thunderbird",
	429: "bison",
	436: "tusker",
	437: "husky",
}

// mockClient is a mock of *sqs.SQS.
type mockClient struct {
	err              error
	sendOutputChan   chan *sqs.SendMessageOutput
	receiveOuputChan chan *sqs.ReceiveMessageOutput
	sendMsgID        string
}

func (mock *mockClient) SendMessageWithContext(ctx context.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	if input != nil && input.MessageBody != nil && *input.MessageBody != "" {
		go func() {
			mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{
						MessageAttributes: input.MessageAttributes,
						Body:              input.MessageBody,
						MessageId:         aws.String(mock.sendMsgID),
					},
				},
			}
		}()
		return &sqs.SendMessageOutput{MessageId: aws.String(mock.sendMsgID)}, nil
	}
	// Add logic to allow context errors.
	for {
		select {
		case d := <-mock.sendOutputChan:
			return d, mock.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (mock *mockClient) ChangeMessageVisibilityWithContext(ctx aws.Context, input *sqs.ChangeMessageVisibilityInput, opts ...request.Option) (*sqs.ChangeMessageVisibilityOutput, error) {
	return nil, nil
}

// TestBadEncode tests if encode errors are handled properly.
func TestBadEncode(t *testing.T) {
	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	mock := &mockClient{
		sendOutputChan: make(chan *sqs.SendMessageOutput),
	}
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return errors.New("err!") },
		func(context.Context, *sqs.Message) (response interface{}, err error) { return struct{}{}, nil },
	)
	errChan := make(chan error, 1)
	var err error
	go func() {
		_, err := pub.Endpoint()(context.Background(), struct{}{})
		errChan <- err

	}()
	select {
	case err = <-errChan:
		break

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for result")
	}
	if err == nil {
		t.Error("expected error")
	}
	if want, have := "err!", err.Error(); want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestBadDecode tests if decode errors are handled properly.
func TestBadDecode(t *testing.T) {
	mock := &mockClient{
		sendOutputChan: make(chan *sqs.SendMessageOutput),
	}
	go func() {
		mock.sendOutputChan <- &sqs.SendMessageOutput{
			MessageId: aws.String("someMsgID"),
		}
	}()

	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		func(context.Context, *sqs.Message) (response interface{}, err error) {
			return struct{}{}, errors.New("err!")
		},
		awssqs.PublisherAfter(func(ctx context.Context, client awssqs.Client, msg *sqs.SendMessageOutput) (context.Context, *sqs.Message, error) {
			// Set the actual response for the request.
			return ctx, &sqs.Message{Body: aws.String("someMsgContent")}, nil
		}),
	)

	var err error
	errChan := make(chan error, 1)
	go func() {
		_, err := pub.Endpoint()(context.Background(), struct{}{})
		errChan <- err
	}()

	select {
	case err = <-errChan:
		break

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for result")
	}

	if err == nil {
		t.Error("expected error")
	}
	if want, have := "err!", err.Error(); want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestPublisherTimeout ensures that the publisher timeout mechanism works.
func TestPublisherTimeout(t *testing.T) {
	sendOutputChan := make(chan *sqs.SendMessageOutput)
	mock := &mockClient{
		sendOutputChan: sendOutputChan,
	}
	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		func(context.Context, *sqs.Message) (response interface{}, err error) {
			return struct{}{}, nil
		},
		awssqs.PublisherTimeout(50*time.Millisecond),
	)

	var err error
	errChan := make(chan error, 1)
	go func() {
		_, err := pub.Endpoint()(context.Background(), struct{}{})
		errChan <- err

	}()

	select {
	case err = <-errChan:
		break

	case <-time.After(1000 * time.Millisecond):
		t.Fatal("timed out waiting for result")
	}

	if err == nil {
		t.Error("expected error")
		return
	}
	if want, have := context.DeadlineExceeded.Error(), err.Error(); want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestSuccessfulPublisher ensures that the publisher mechanisms work.
func TestSuccessfulPublisher(t *testing.T) {
	mockReq := testReq{437}
	mockRes := testRes{
		Squadron: mockReq.Squadron,
		Name:     names[mockReq.Squadron],
	}
	b, err := json.Marshal(mockRes)
	if err != nil {
		t.Fatal(err)
	}
	mock := &mockClient{
		sendOutputChan: make(chan *sqs.SendMessageOutput),
		sendMsgID:      "someMsgID",
	}
	go func() {
		mock.sendOutputChan <- &sqs.SendMessageOutput{
			MessageId: aws.String("someMsgID"),
		}
	}()

	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		awssqs.EncodeJSONRequest,
		func(_ context.Context, msg *sqs.Message) (interface{}, error) {
			response := testRes{}
			err := json.Unmarshal([]byte(*msg.Body), &response)
			return response, err
		},
		awssqs.PublisherAfter(func(ctx context.Context, client awssqs.Client, msg *sqs.SendMessageOutput) (context.Context, *sqs.Message, error) {
			// Sets the actual response for the request.
			if *msg.MessageId == "someMsgID" {
				return ctx, &sqs.Message{Body: aws.String(string(b))}, nil
			}
			return nil, nil, fmt.Errorf("Did not receive expected SendMessageOutput")
		}),
	)
	var res testRes
	var ok bool
	resChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	go func() {
		res, err := pub.Endpoint()(context.Background(), mockReq)
		if err != nil {
			errChan <- err
		} else {
			resChan <- res
		}
	}()

	select {
	case response := <-resChan:
		res, ok = response.(testRes)
		if !ok {
			t.Error("failed to assert endpoint response type")
		}
		break

	case err = <-errChan:
		break

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for result")
	}

	if err != nil {
		t.Fatal(err)
	}
	if want, have := mockRes.Name, res.Name; want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestSuccessfulPublisherNoResponse ensures that the publisher response mechanism works.
func TestSuccessfulPublisherNoResponse(t *testing.T) {
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
		sendMsgID:        "someMsgID",
	}

	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		awssqs.EncodeJSONRequest,
		awssqs.NoResponseDecode,
	)
	var err error
	errChan := make(chan error, 1)
	finishChan := make(chan bool, 1)
	go func() {
		_, err := pub.Endpoint()(context.Background(), struct{}{})
		if err != nil {
			errChan <- err
		} else {
			finishChan <- true
		}
	}()

	select {
	case <-finishChan:
		break
	case err = <-errChan:
		t.Errorf("unexpected error %s", err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for result")
	}
}

// TestPublisherWithBefore adds a PublisherBefore function that adds a message attribute.
// This test ensures that the the before functions work as expected.
func TestPublisherWithBefore(t *testing.T) {
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
		sendMsgID:        "someMsgID",
	}

	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewPublisher(
		mock,
		queueURL,
		responseQueueURL,
		awssqs.EncodeJSONRequest,
		awssqs.NoResponseDecode,
		awssqs.PublisherBefore(func(c context.Context, s *sqs.SendMessageInput) context.Context {
			if s.MessageAttributes == nil {
				s.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
			}
			s.MessageAttributes["responseQueueURL"] = &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: &responseQueueURL,
			}
			return c
		}),
	)
	var err error
	errChan := make(chan error, 1)
	go func() {
		_, err := pub.Endpoint()(context.Background(), struct{}{})
		if err != nil {
			errChan <- err
		}
	}()

	want := sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: &responseQueueURL,
	}

	select {
	case receiveOutput := <-mock.receiveOuputChan:
		if len(receiveOutput.Messages) != 1 {
			t.Errorf("published %d messages instead of 1", len(receiveOutput.Messages))
		}
		if have, exists := receiveOutput.Messages[0].MessageAttributes["responseQueueURL"]; !exists {
			t.Errorf("expected MessageAttributes responseQueueURL not found")
		} else if *have.StringValue != responseQueueURL || *have.DataType != "String" {
			t.Errorf("want %s, have %s", want, *have)
		}
		break
	case err = <-errChan:
		t.Errorf("unexpected error %s", err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for result")
	}
}
