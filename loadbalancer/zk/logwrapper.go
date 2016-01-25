package zk

import (
	"fmt"

	"github.com/samuel/go-zookeeper/zk"

	"github.com/go-kit/kit/log"
)

// wrapLogger wraps a go-kit logger so we can use it as the logging service for
// the ZooKeeper library (which expects a Printf method to be available)
type wrapLogger struct {
	log.Logger
}

func (logger wrapLogger) Printf(str string, vars ...interface{}) {
	logger.Log("msg", fmt.Sprintf(str, vars...))
}

// withLogger replaces the ZooKeeper library's default logging service for our
// own go-kit logger
func withLogger(logger log.Logger) func(c *zk.Conn) {
	return func(c *zk.Conn) {
		c.SetLogger(wrapLogger{logger})
	}
}
