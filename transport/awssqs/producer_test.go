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
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
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
	sqsiface.SQSAPI
	err              error
	sendOutputChan   chan *sqs.SendMessageOutput
	receiveOuputChan chan *sqs.ReceiveMessageOutput
	sendMsgID        string
	deleteError      error
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
	mock := &mockClient{
		sendOutputChan: make(chan *sqs.SendMessageOutput),
	}
	pub := awssqs.NewProducer(
		mock,
		queueURL,
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
	pub := awssqs.NewProducer(
		mock,
		queueURL,
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		func(context.Context, *sqs.Message) (response interface{}, err error) {
			return struct{}{}, errors.New("err!")
		},
		awssqs.ProducerAfter(func(ctx context.Context, _ sqsiface.SQSAPI, msg *sqs.SendMessageOutput) (context.Context, *sqs.Message, error) {
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

// TestProducerTimeout ensures that the producer timeout mechanism works.
func TestProducerTimeout(t *testing.T) {
	sendOutputChan := make(chan *sqs.SendMessageOutput)
	mock := &mockClient{
		sendOutputChan: sendOutputChan,
	}
	queueURL := "someURL"
	pub := awssqs.NewProducer(
		mock,
		queueURL,
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		func(context.Context, *sqs.Message) (response interface{}, err error) {
			return struct{}{}, nil
		},
		awssqs.ProducerTimeout(50*time.Millisecond),
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

// TestSuccessfulProducer ensures that the producer mechanisms work.
func TestSuccessfulProducer(t *testing.T) {
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
	pub := awssqs.NewProducer(
		mock,
		queueURL,
		awssqs.EncodeJSONRequest,
		func(_ context.Context, msg *sqs.Message) (interface{}, error) {
			response := testRes{}
			err := json.Unmarshal([]byte(*msg.Body), &response)
			return response, err
		},
		awssqs.ProducerAfter(func(ctx context.Context, _ sqsiface.SQSAPI, msg *sqs.SendMessageOutput) (context.Context, *sqs.Message, error) {
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

// TestSuccessfulProducerNoResponse ensures that the producer response mechanism works.
func TestSuccessfulProducerNoResponse(t *testing.T) {
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
		sendMsgID:        "someMsgID",
	}

	queueURL := "someURL"
	pub := awssqs.NewProducer(
		mock,
		queueURL,
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

// TestProducerWithBefore adds a ProducerBefore function that adds responseQueueURL to context,
// and another on that adds it as a message attribute to outgoing message.
// This test ensures that setting multiple before functions work as expected
// and that SetProducerResponseQueueURL works as expected.
func TestProducerWithBefore(t *testing.T) {
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
		sendMsgID:        "someMsgID",
	}

	queueURL := "someURL"
	responseQueueURL := "someOtherURL"
	pub := awssqs.NewProducer(
		mock,
		queueURL,
		awssqs.EncodeJSONRequest,
		awssqs.NoResponseDecode,
		awssqs.ProducerBefore(awssqs.SetProducerResponseQueueURL(responseQueueURL)),
		awssqs.ProducerBefore(func(c context.Context, s *sqs.SendMessageInput) context.Context {
			responseQueueURL := c.Value(awssqs.ContextKeyResponseQueueURL).(string)
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
