package nats

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-kit/kit/endpoint"
	"github.com/nats-io/go-nats"
	"time"
	"os"

	log "github.com/sirupsen/logrus"

)

//Client wraps a nats connection and provides a subject that implements it
type Client struct {

	serviceName string
	subject     string
	enc         EncodeRequestFunc
	dec         DecodeRequestFunc
	before      []ClientRequestFunc
	after       []ClientResponseFunc
	natsReply   reflect.Type
	logger log.Logger
}

// NewClient constructs a usable Client for a single remote endpoint.

func NewClient(
	serviceName string,
	subject string,
	enc EncodeRequestFunc,
	dec DecodeRequestFunc,
	natsReply interface{},
	options ...ClientOption,
) *Client {
	c := &Client{
		subject: fmt.Sprintf("/%s/%s", serviceName, subject),
		enc:     enc,
		dec:     dec,
		// We are using reflect.Indirect here to allow both reply structs and
		// pointers to these reply structs. New consumers of the client should
		// use structs directly, while existing consumers will not break if they
		// remain to use pointers to structs.
		natsReply: reflect.TypeOf(
			reflect.Indirect(
				reflect.ValueOf(natsReply),
			).Interface(),
		),
		before: []ClientRequestFunc{},
		after:  []ClientResponseFunc{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// ClientBefore sets the RequestFuncs that are applied to the outgoing gRPC
// request before it's invoked.
func ClientBefore(before ...ClientRequestFunc) ClientOption {
	return func(c *Client) { c.before = append(c.before, before...) }
}

// ClientAfter sets the ClientResponseFuncs that are applied to the incoming
// gRPC response prior to it being decoded. This is useful for obtaining
// response metadata and adding onto the context prior to decoding.
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(c *Client) { c.after = append(c.after, after...) }
}

// Endpoint returns a usable endpoint that will invoke the nats specified by the
// client.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		urls := fmt.Sprint(os.Getenv("NATS_SERVER"), ", ", fmt.Sprintf("nats://%v:%v", os.Getenv("NATS_SERVICE_HOST"), os.Getenv("NATS_SERVICE_PORT")))

		c.logger.Info(urls)

		nc, err := nats.Connect(urls)
		if err != nil {
			c.logger.Error("Can't connect: %v\n", err)
		}
		defer nc.Close()

		var msg *nats.Msg
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		req, err := c.enc(ctx,request)
		if err != nil {
			return nil, err
		}

		if msg, err = nc.Request(c.subject, req, time.Second * 10); err != nil {
			return nil, err
		}

		//for _, f := range c.after {
		//	ctx = f(ctx, header, trailer)
		//}

		response, err := c.dec(ctx, msg)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}
