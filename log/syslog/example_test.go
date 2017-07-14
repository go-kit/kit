package syslog_test

import (
	gosyslog "log/syslog"
	"github.com/go-kit/kit/log/syslog"
	"github.com/go-kit/kit/log"
	"fmt"
	"github.com/go-kit/kit/log/level"
)


func ExampleNewLogger_defaultPrioritySelector() {
	// Normal syslog writer
	w, err := gosyslog.New(gosyslog.LOG_INFO, "experiment")
	if err != nil {
		fmt.Println(err)
		return
	}

	// syslog logger with logfmt formatting
	logger := syslog.NewSyslogLogger(w, log.NewLogfmtLogger)
	logger.Log("msg", "info because of default")
	logger.Log(level.Key(), level.DebugValue(), "msg", "debug because of explicit level")
}