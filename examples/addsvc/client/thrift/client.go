package thrift

import (
	"io"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/add"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

// New returns a stateful factory for Sum and Concat Endpoints
func New(protocol string, bufferSize int, framed bool, logger log.Logger) client {
	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		panic("invalid protocol")
	}

	var transportFactory thrift.TTransportFactory
	if bufferSize > 0 {
		transportFactory = thrift.NewTBufferedTransportFactory(bufferSize)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}
	if framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	return client{transportFactory, protocolFactory, logger}
}

type client struct {
	thrift.TTransportFactory
	thrift.TProtocolFactory
	log.Logger
}

// SumEndpointFactory transforms host:port strings into Endpoints.
func (c client) SumEndpoint(instance string) (endpoint.Endpoint, io.Closer, error) {
	transportSocket, err := thrift.NewTSocket(instance)
	if err != nil {
		c.Logger.Log("during", "thrift.NewTSocket", "err", err)
		return nil, nil, err
	}
	trans := c.TTransportFactory.GetTransport(transportSocket)

	if err := trans.Open(); err != nil {
		c.Logger.Log("during", "thrift transport.Open", "err", err)
		return nil, nil, err
	}
	cli := thriftadd.NewAddServiceClientFactory(trans, c.TProtocolFactory)

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		sumRequest := request.(server.SumRequest)
		reply, err := cli.Sum(int64(sumRequest.A), int64(sumRequest.B))
		if err != nil {
			return server.SumResponse{}, err
		}
		return server.SumResponse{V: int(reply.Value)}, nil
	}, trans, nil
}

// ConcatEndpointFactory transforms host:port strings into Endpoints.
func (c client) ConcatEndpoint(instance string) (endpoint.Endpoint, io.Closer, error) {
	transportSocket, err := thrift.NewTSocket(instance)
	if err != nil {
		c.Logger.Log("during", "thrift.NewTSocket", "err", err)
		return nil, nil, err
	}
	trans := c.TTransportFactory.GetTransport(transportSocket)

	if err := trans.Open(); err != nil {
		c.Logger.Log("during", "thrift transport.Open", "err", err)
		return nil, nil, err
	}
	cli := thriftadd.NewAddServiceClientFactory(trans, c.TProtocolFactory)

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		concatRequest := request.(server.ConcatRequest)
		reply, err := cli.Concat(concatRequest.A, concatRequest.B)
		if err != nil {
			return server.ConcatResponse{}, err
		}
		return server.ConcatResponse{V: reply.Value}, nil
	}, trans, nil
}
