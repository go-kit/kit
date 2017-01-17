package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"
	"time"
)

func TestStdlibWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(log.LstdFlags)
	logger := NewLogfmtLogger(StdlibWriter{})
	logger.Log("key", "val")
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	if want, have := timestamp+" key=val\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestStdlibAdapterUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogfmtLogger(buf)

	for _, test := range []struct {
		writer io.Writer
		now    time.Time
		flag   int
	}{
		{NewStdlibAdapter(logger), time.Now(), 0},
		{NewStdlibAdapter(logger, UTC()), time.Now().UTC(), log.LUTC},
	} {
		stdlog := log.New(test.writer, "", 0)
		now := test.now.Format(time.RFC3339)

		for flag, want := range map[int]string{
			0:                                      "msg=hello\n",
			log.Ldate:                              "ts=" + now + " msg=hello\n",
			log.Ltime:                              "ts=" + now + " msg=hello\n",
			log.Ldate | log.Ltime:                  "ts=" + now + " msg=hello\n",
			log.Lshortfile:                         "caller=stdlib_test.go:50 msg=hello\n",
			log.Lshortfile | log.Ldate:             "ts=" + now + " caller=stdlib_test.go:50 msg=hello\n",
			log.Lshortfile | log.Ldate | log.Ltime: "ts=" + now + " caller=stdlib_test.go:50 msg=hello\n",
		} {
			buf.Reset()
			stdlog.SetFlags(flag | test.flag)
			stdlog.Print("hello")
			if have := buf.String(); want != have {
				t.Errorf("flag=%d, now=%v: want %#v, have %#v", flag, test.now, want, have)
			}
		}
	}
}

func TestStdLibAdapterExtraction(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogfmtLogger(buf)
	writer := NewStdlibAdapter(logger)
	now := time.Now()
	stdday := now.Format(formatStdDate)
	stdsec := now.Format(formatStdTime)
	stdmic := now.Format(formatStdTime + formatStdMicro)
	fmtsec := now.Format(time.RFC3339)
	fmtmic := now.Format(formatRFC3339Micro)

	for input, want := range map[string]string{
		"hello":                                          "msg=hello\n",
		stdday + ": hello":                               "ts=" + fmtsec + " msg=hello\n",
		stdday + " " + stdsec + ": hello":                "ts=" + fmtsec + " msg=hello\n",
		stdsec + ": hello":                               "ts=" + fmtsec + " msg=hello\n",
		stdday + " " + stdmic + ": hello":                "ts=" + fmtmic + " msg=hello\n",
		stdday + " " + stdmic + " /a/b/c/d.go:23: hello": "ts=" + fmtmic + " caller=/a/b/c/d.go:23 msg=hello\n",
		stdmic + " /a/b/c/d.go:23: hello":                "ts=" + fmtmic + " caller=/a/b/c/d.go:23 msg=hello\n",
		stdday + " " + stdsec + " /a/b/c/d.go:23: hello": "ts=" + fmtsec + " caller=/a/b/c/d.go:23 msg=hello\n",
		stdday + " /a/b/c/d.go:23: hello":                "ts=" + fmtsec + " caller=/a/b/c/d.go:23 msg=hello\n",
		"/a/b/c/d.go:23: hello":                          "caller=/a/b/c/d.go:23 msg=hello\n",
	} {
		buf.Reset()
		fmt.Fprint(writer, input)
		if have := buf.String(); want != have {
			t.Errorf("%q: want %#v, have %#v", input, want, have)
		}
	}
}

func TestStdlibAdapterSubexps(t *testing.T) {
	for input, wantMap := range map[string]map[string]string{
		"hello world": {
			"date": "",
			"time": "",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23: hello world": {
			"date": "2009/01/23",
			"time": "",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "",
			"msg":  "hello world",
		},
		"01:23:23: hello world": {
			"date": "",
			"time": "01:23:23",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123 /a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"01:23:23.123123 /a/b/c/d.go:23: hello world": {
			"date": "",
			"time": "01:23:23.123123",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23 /a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 /a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"/a/b/c/d.go:23: hello world": {
			"date": "",
			"time": "",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123 C:/a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "C:/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"01:23:23.123123 C:/a/b/c/d.go:23: hello world": {
			"date": "",
			"time": "01:23:23.123123",
			"file": "C:/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23 C:/a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "C:/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 C:/a/b/c/d.go:23: hello world": {
			"date": "2009/01/23",
			"time": "",
			"file": "C:/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"C:/a/b/c/d.go:23: hello world": {
			"date": "",
			"time": "",
			"file": "C:/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123 C:/a/b/c/d.go:23: :.;<>_#{[]}\"\\": {
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "C:/a/b/c/d.go:23",
			"msg":  ":.;<>_#{[]}\"\\",
		},
		"01:23:23.123123 C:/a/b/c/d.go:23: :.;<>_#{[]}\"\\": {
			"date": "",
			"time": "01:23:23.123123",
			"file": "C:/a/b/c/d.go:23",
			"msg":  ":.;<>_#{[]}\"\\",
		},
		"2009/01/23 01:23:23 C:/a/b/c/d.go:23: :.;<>_#{[]}\"\\": {
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "C:/a/b/c/d.go:23",
			"msg":  ":.;<>_#{[]}\"\\",
		},
		"2009/01/23 C:/a/b/c/d.go:23: :.;<>_#{[]}\"\\": {
			"date": "2009/01/23",
			"time": "",
			"file": "C:/a/b/c/d.go:23",
			"msg":  ":.;<>_#{[]}\"\\",
		},
		"C:/a/b/c/d.go:23: :.;<>_#{[]}\"\\": {
			"date": "",
			"time": "",
			"file": "C:/a/b/c/d.go:23",
			"msg":  ":.;<>_#{[]}\"\\",
		},
	} {
		haveMap := subexps([]byte(input))
		for key, want := range wantMap {
			if have := haveMap[key]; want != have {
				t.Errorf("%q: %q: want %q, have %q", input, key, want, have)
			}
		}
	}
}
