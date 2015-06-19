package thrift

import (
	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/context"

	thriftadd "github.com/go-kit/kit/addsvc/_thrift/gen-go/add"
	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
)

// NewClient takes a Thrift Transport and protocol pair which should point to
// an instance of an addsvc. It returns an endpoint that wraps and invokes
// that transport.
func NewClient(transport thrift.TTransport, input, output thrift.TProtocol) endpoint.Endpoint {
	client := thriftadd.NewAddServiceClientProtocol(transport, input, output)
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		addReq, ok := request.(*reqrep.AddRequest)
		if !ok {
			return nil, endpoint.ErrBadCast
		}
		reply, err := client.Add(addReq.A, addReq.B)
		if err != nil {
			return nil, err
		}
		return &reqrep.AddResponse{V: reply.Value}, nil
	}
}
