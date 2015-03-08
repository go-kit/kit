package log_test

import (
	"reflect"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestLogger(t *testing.T) {
	want := [][]interface{}{
		{"msg", "Hello, world!"},
	}

	got := [][]interface{}{}
	l := sliceLogger(&got)

	for _, d := range want {
		l.Log(d...)
	}

	for i := range want {
		if g, w := got[i], want[i]; !reflect.DeepEqual(g, w) {
			t.Errorf("\n got %v\nwant %v", g, w)
		}
	}
}

func TestKeyValue(t *testing.T) {
	now := log.Now()

	data := []struct {
		in  []interface{}
		out []interface{}
	}{
		{
			in:  []interface{}{log.Debug},
			out: []interface{}{log.LvlKey, log.Debug},
		},
		{
			in:  []interface{}{log.Info},
			out: []interface{}{log.LvlKey, log.Info},
		},
		{
			in:  []interface{}{log.Warn},
			out: []interface{}{log.LvlKey, log.Warn},
		},
		{
			in:  []interface{}{log.Error},
			out: []interface{}{log.LvlKey, log.Error},
		},
		{
			in:  []interface{}{log.Crit},
			out: []interface{}{log.LvlKey, log.Crit},
		},
		{
			in:  []interface{}{now},
			out: []interface{}{log.TimeKey, now.Value()},
		},
	}

	got := [][]interface{}{}
	l := sliceLogger(&got)

	for _, d := range data {
		l.Log(d.in...)
	}

	for i := range data {
		if g, w := got[i], data[i].out; !reflect.DeepEqual(g, w) {
			t.Errorf("\n got %v\nwant %v", g, w)
		}
	}
}

func sliceLogger(s *[][]interface{}) log.Logger {
	l := log.New()
	*s = [][]interface{}{}
	l.SetHandler(log.HandlerFunc(func(keyvals ...interface{}) error {
		*s = append(*s, keyvals)
		return nil
	}))
	return l
}
