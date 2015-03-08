package log_test

import (
	"os"
	"time"

	"github.com/peterbourgon/gokit/log"
)

func Example_log() {
	h := log.Writer(os.Stdout, log.JsonEncoder())
	h = log.ReplaceKeys(h, log.TimeKey, "t", log.LvlKey, "lvl")

	// Use a static time so it will always match the output below
	now := time.Date(2015, time.March, 7, 20, 12, 33, 0, time.UTC)

	// A real program would add timestamps in the handler chain
	// h = log.AddTimestamp(h)

	l := log.New()
	l.SetHandler(h)

	l.Log(log.Info, log.TimeKey, now, "msg", "Hello, world!")
	// Output:
	// {"lvl":"info","msg":"Hello, world!","t":"2015-03-07T20:12:33Z"}
}

func Example_logContext() {
	h := log.Writer(os.Stdout, log.JsonEncoder())
	h = log.ReplaceKeys(h, log.LvlKey, "lvl")

	// Create a logger with some default KVs
	l := log.New("host", "hal9000")
	l.SetHandler(h)

	// Create another new logger from the first, adding another KV.
	l2 := l.New("url", "/index.html")

	l2.Log(log.Info, "msg", "Hello, world!")
	// Output:
	// {"host":"hal9000","lvl":"info","msg":"Hello, world!","url":"/index.html"}
}
