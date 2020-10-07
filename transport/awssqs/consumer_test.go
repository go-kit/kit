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
	"github.com/pborman/uuid"
)

var (
	errTypeAssertion = errors.New("type assertion error")
)

func (mock *mockClient) ReceiveMessageWithContext(ctx context.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	// Add logic to allow context errors
	for {
		select {
		case d := <-mock.receiveOuputChan:
			return d, mock.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// TestConsumerBadDecode checks if decoder errors are handled properly.
func TestConsumerBadDecode(t *testing.T) {
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      aws.String("MessageBody"),
					MessageId: aws.String("fakeMsgID"),
				},
			},
		}
	}()
	errEncoder := awssqs.ConsumerErrorEncoder(func(ctx context.Context, err error, req *sqs.Message, sqsClient awssqs.Client) {
		publishError := sqsError{
			Err:   err.Error(),
			MsgID: *req.MessageId,
		}
		payload, _ := json.Marshal(publishError)

		sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
			MessageBody: aws.String(string(payload)),
		})
	})
	consumer := awssqs.NewConsumer(mock,
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, *sqs.Message) (interface{}, error) { return nil, errors.New("err!") },
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		queueURL,
		errEncoder,
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)

	consumer.Consume(context.Background(), &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	res, err := decodeConsumerError(receiveOutput)
	if err != nil {
		t.Fatal(err)
	}
	if want, have := "err!", res.Err; want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestConsumerBadEndpoint checks if endpoint errors are handled properly.
func TestConsumerBadEndpoint(t *testing.T) {
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      aws.String("MessageBody"),
					MessageId: aws.String("fakeMsgID"),
				},
			},
		}
	}()
	errEncoder := awssqs.ConsumerErrorEncoder(func(ctx context.Context, err error, req *sqs.Message, sqsClient awssqs.Client) {
		publishError := sqsError{
			Err:   err.Error(),
			MsgID: *req.MessageId,
		}
		payload, _ := json.Marshal(publishError)

		sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
			MessageBody: aws.String(string(payload)),
		})
	})
	consumer := awssqs.NewConsumer(mock,
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errors.New("err!") },
		func(context.Context, *sqs.Message) (interface{}, error) { return nil, nil },
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return nil },
		queueURL,
		errEncoder,
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)

	consumer.Consume(context.Background(), &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	res, err := decodeConsumerError(receiveOutput)
	if err != nil {
		t.Fatal(err)
	}
	if want, have := "err!", res.Err; want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestConsumerBadEncoder checks if encoder errors are handled properly.
func TestConsumerBadEncoder(t *testing.T) {
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      aws.String("MessageBody"),
					MessageId: aws.String("fakeMsgID"),
				},
			},
		}
	}()
	errEncoder := awssqs.ConsumerErrorEncoder(func(ctx context.Context, err error, req *sqs.Message, sqsClient awssqs.Client) {
		publishError := sqsError{
			Err:   err.Error(),
			MsgID: *req.MessageId,
		}
		payload, _ := json.Marshal(publishError)

		sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
			MessageBody: aws.String(string(payload)),
		})
	})
	consumer := awssqs.NewConsumer(mock,
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, *sqs.Message) (interface{}, error) { return nil, nil },
		func(context.Context, *sqs.SendMessageInput, interface{}) error { return errors.New("err!") },
		queueURL,
		errEncoder,
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)

	consumer.Consume(context.Background(), &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	res, err := decodeConsumerError(receiveOutput)
	if err != nil {
		t.Fatal(err)
	}
	if want, have := "err!", res.Err; want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

// TestConsumerSuccess checks if consumer responds correctly to message.
func TestConsumerSuccess(t *testing.T) {
	obj := testReq{
		Squadron: 436,
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      aws.String(string(b)),
					MessageId: aws.String("fakeMsgID"),
				},
			},
		}
	}()
	consumer := awssqs.NewConsumer(mock,
		testEndpoint,
		testReqDecoderfunc,
		awssqs.EncodeJSONResponse,
		queueURL,
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)

	consumer.Consume(context.Background(), &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	res, err := decodeResponse(receiveOutput)
	if err != nil {
		t.Fatal(err)
	}
	want := testRes{
		Squadron: 436,
		Name:     "tusker",
	}
	if have := res; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

// TestConsumerSuccessNoReply checks if consumer processes correctly message
// without sending response.
func TestConsumerSuccessNoReply(t *testing.T) {
	obj := testReq{
		Squadron: 436,
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body:      aws.String(string(b)),
					MessageId: aws.String("fakeMsgID"),
				},
			},
		}
	}()
	consumer := awssqs.NewConsumer(mock,
		testEndpoint,
		testReqDecoderfunc,
		awssqs.EncodeJSONResponse,
		queueURL,
	)

	consumer.Consume(context.Background(), &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		t.Errorf("received output when none was expected, have %v", receiveOutput)
		return

	case <-time.After(200 * time.Millisecond):
		// As expected, we did not receive any response from consumer
		return
	}
}

// TestConsumerBeforeFilterMessages checks if consumer before is called as expected.
// Here before is used to filter messages before processing
func TestConsumerBeforeFilterMessages(t *testing.T) {
	obj1 := testReq{
		Squadron: 436,
	}
	b1, _ := json.Marshal(obj1)
	obj2 := testReq{
		Squadron: 4,
	}
	b2, _ := json.Marshal(obj2)
	obj3 := testReq{
		Squadron: 1,
	}
	b3, _ := json.Marshal(obj3)
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	expectedMsgs := []*sqs.Message{
		{
			Body:      aws.String(string(b1)),
			MessageId: aws.String("fakeMsgID1"),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"recipient": {
					DataType:    aws.String("String"),
					StringValue: aws.String("me"),
				},
			},
		},
		{
			Body:      aws.String(string(b2)),
			MessageId: aws.String("fakeMsgID2"),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"recipient": {
					DataType:    aws.String("String"),
					StringValue: aws.String("not me"),
				},
			},
		},
		{
			Body:      aws.String(string(b3)),
			MessageId: aws.String("fakeMsgID3"),
		},
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: expectedMsgs,
		}
	}()
	type ctxKey struct {
		key string
	}
	consumer := awssqs.NewConsumer(mock,
		testEndpoint,
		testReqDecoderfunc,
		awssqs.EncodeJSONResponse,
		queueURL,
		awssqs.ConsumerBefore(func(ctx context.Context, msgs *[]*sqs.Message) context.Context {
			// delete a message that is not destined to the consumer
			msgsCopy := *msgs
			for index, msg := range *msgs {
				if recipient, exists := msg.MessageAttributes["recipient"]; !exists || *recipient.StringValue != "me" {
					msgsCopy = append(msgsCopy[:index], msgsCopy[index:]...)
				}
			}
			*msgs = msgsCopy
			return ctx
		}),
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)
	ctx := context.Background()
	consumer.Consume(ctx, &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	res, err := decodeResponse(receiveOutput)
	if err != nil {
		t.Fatal(err)
	}
	want := testRes{
		Squadron: 436,
		Name:     "tusker",
	}
	if have := res; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
	// Try fetching responses again
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		t.Errorf("received second output when only one was expected, have %v", receiveOutput)
		return

	case <-time.After(200 * time.Millisecond):
		// As expected, we did not receive a second response from consumer
		return
	}
}

// TestConsumerAfter checks if consumer after is called as expected.
// Here after is used to transfer some info from received message in response.
func TestConsumerAfter(t *testing.T) {
	obj1 := testReq{
		Squadron: 436,
	}
	b1, _ := json.Marshal(obj1)
	queueURL := "someURL"
	mock := &mockClient{
		sendOutputChan:   make(chan *sqs.SendMessageOutput),
		receiveOuputChan: make(chan *sqs.ReceiveMessageOutput),
	}
	correlationID := uuid.NewRandom().String()
	expectedMsgs := []*sqs.Message{
		{
			Body:      aws.String(string(b1)),
			MessageId: aws.String("fakeMsgID1"),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"correlationID": {
					DataType:    aws.String("String"),
					StringValue: &correlationID,
				},
			},
		},
	}
	go func() {
		mock.receiveOuputChan <- &sqs.ReceiveMessageOutput{
			Messages: expectedMsgs,
		}
	}()
	type ctxKey struct {
		key string
	}
	consumer := awssqs.NewConsumer(mock,
		testEndpoint,
		testReqDecoderfunc,
		awssqs.EncodeJSONResponse,
		queueURL,
		awssqs.ConsumerAfter(func(ctx context.Context, msg *sqs.Message, resp *sqs.SendMessageInput, leftMsgs *[]*sqs.Message) context.Context {
			if correlationIDAttribute, exists := msg.MessageAttributes["correlationID"]; exists {
				if resp.MessageAttributes == nil {
					resp.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
				}
				resp.MessageAttributes["correlationID"] = &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: correlationIDAttribute.StringValue,
				}
			}
			return ctx
		}),
		awssqs.ConsumerWantReplyFunc(func(context.Context, *sqs.Message) bool { return true }),
	)
	ctx := context.Background()
	consumer.Consume(ctx, &sqs.ReceiveMessageInput{})

	var receiveOutput *sqs.ReceiveMessageOutput
	select {
	case receiveOutput = <-mock.receiveOuputChan:
		break

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for publishing")
	}
	if len(receiveOutput.Messages) != 1 {
		t.Errorf("received %d messages instead of 1", len(receiveOutput.Messages))
	}
	if correlationIDAttribute, exists := receiveOutput.Messages[0].MessageAttributes["correlationID"]; exists {
		if have := correlationIDAttribute.StringValue; *have != correlationID {
			t.Errorf("have %s, want %s", *have, correlationID)
		}
	} else {
		t.Errorf("expected message attribute with key correlationID in response, but it was not found")
	}
}

type sqsError struct {
	Err   string `json:"err"`
	MsgID string `json:"msgID"`
}

func decodeConsumerError(receiveOutput *sqs.ReceiveMessageOutput) (sqsError, error) {
	receivedError := sqsError{}
	err := json.Unmarshal([]byte(*receiveOutput.Messages[0].Body), &receivedError)
	return receivedError, err
}

func testEndpoint(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(testReq)
	if !ok {
		return nil, errTypeAssertion
	}
	name, prs := names[req.Squadron]
	if !prs {
		return nil, errors.New("unknown squadron name")
	}
	res := testRes{
		Squadron: req.Squadron,
		Name:     name,
	}
	return res, nil
}

func testReqDecoderfunc(_ context.Context, msg *sqs.Message) (interface{}, error) {
	var obj testReq
	err := json.Unmarshal([]byte(*msg.Body), &obj)
	return obj, err
}

func decodeResponse(receiveOutput *sqs.ReceiveMessageOutput) (interface{}, error) {
	if len(receiveOutput.Messages) != 1 {
		return nil, fmt.Errorf("Error : received %d messages instead of 1", len(receiveOutput.Messages))
	}
	resp := testRes{}
	err := json.Unmarshal([]byte(*receiveOutput.Messages[0].Body), &resp)
	return resp, err
}
