package netrpc

import (
	"net/rpc"

	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/log"
)

// New returns an AddService that's backed by the URL. baseurl will have its
// scheme and hostport used, but its path will be overwritten. If client is
// nil, http.DefaultClient will be used.
func New(cli *rpc.Client, logger log.Logger) server.AddService {
	return client{cli, logger}
}

type client struct {
	*rpc.Client
	log.Logger
}

func (c client) Sum(a, b int) int {
	var reply server.SumResponse
	if err := c.Client.Call("addsvc.Sum", server.SumRequest{A: a, B: b}, &reply); err != nil {
		_ = c.Logger.Log("err", err)
		return 0
	}
	return reply.V
}

func (c client) Concat(a, b string) string {
	var reply server.ConcatResponse
	if err := c.Client.Call("addsvc.Concat", server.ConcatRequest{A: a, B: b}, &reply); err != nil {
		_ = c.Logger.Log("err", err)
		return ""
	}
	return reply.V
}
