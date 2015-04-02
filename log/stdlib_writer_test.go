package log

import (
	"bytes"
	"fmt"
	"testing"
)

func TestStdLibWriterExtraction(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewPrefixLogger(buf)
	writer := NewStdlibWriter(logger)
	for input, want := range map[string]string{
		"hello":                                            "msg=hello\n",
		"2009/01/23: hello":                                "ts=2009/01/23 msg=hello\n",
		"2009/01/23 01:23:23: hello":                       "ts=2009/01/23 01:23:23 msg=hello\n",
		"01:23:23: hello":                                  "ts=01:23:23 msg=hello\n",
		"2009/01/23 01:23:23.123123: hello":                "ts=2009/01/23 01:23:23.123123 msg=hello\n",
		"2009/01/23 01:23:23.123123 /a/b/c/d.go:23: hello": "ts=2009/01/23 01:23:23.123123 file=/a/b/c/d.go:23 msg=hello\n",
		"01:23:23.123123 /a/b/c/d.go:23: hello":            "ts=01:23:23.123123 file=/a/b/c/d.go:23 msg=hello\n",
		"2009/01/23 01:23:23 /a/b/c/d.go:23: hello":        "ts=2009/01/23 01:23:23 file=/a/b/c/d.go:23 msg=hello\n",
		"2009/01/23 /a/b/c/d.go:23: hello":                 "ts=2009/01/23 file=/a/b/c/d.go:23 msg=hello\n",
		"/a/b/c/d.go:23: hello":                            "file=/a/b/c/d.go:23 msg=hello\n",
	} {
		buf.Reset()
		fmt.Fprintf(writer, input)
		if have := buf.String(); want != have {
			t.Errorf("%q: want %#v, have %#v", input, want, have)
		}
	}
}

func TestStdlibWriterSubexps(t *testing.T) {
	for input, wantMap := range map[string]map[string]string{
		"hello world": map[string]string{
			"date": "",
			"time": "",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "",
			"msg":  "hello world",
		},
		"01:23:23: hello world": map[string]string{
			"date": "",
			"time": "01:23:23",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23.123123 /a/b/c/d.go:23: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "01:23:23.123123",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"01:23:23.123123 /a/b/c/d.go:23: hello world": map[string]string{
			"date": "",
			"time": "01:23:23.123123",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 01:23:23 /a/b/c/d.go:23: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "01:23:23",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"2009/01/23 /a/b/c/d.go:23: hello world": map[string]string{
			"date": "2009/01/23",
			"time": "",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
		},
		"/a/b/c/d.go:23: hello world": map[string]string{
			"date": "",
			"time": "",
			"file": "/a/b/c/d.go:23",
			"msg":  "hello world",
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
