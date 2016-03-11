package thrift

import (
	"github.com/go-kit/kit/examples/addsvc/server"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/add"
	"github.com/go-kit/kit/log"
)

// New returns an AddService that's backed by the Thrift client.
func New(cli *thriftadd.AddServiceClient, logger log.Logger) server.AddService {
	return &client{cli, logger}
}

type client struct {
	*thriftadd.AddServiceClient
	log.Logger
}

func (c client) Sum(a, b int) int {
	reply, err := c.AddServiceClient.Sum(int64(a), int64(b))
	if err != nil {
		c.Logger.Log("err", err)
		return 0
	}
	return int(reply.Value)
}

func (c client) Concat(a, b string) string {
	reply, err := c.AddServiceClient.Concat(a, b)
	if err != nil {
		c.Logger.Log("err", err)
		return ""
	}
	return reply.Value
}
