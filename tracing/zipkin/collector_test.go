package zipkin_test

import (
	"fmt"
	"testing"

	"github.com/go-kit/kit/tracing/zipkin"
)

var s *zipkin.Span = zipkin.NewSpan("203.0.113.10:1234", "service1", "avg", 123, 456, 0)

func TestNopCollector(t *testing.T) {
	c := zipkin.NopCollector{}
	if err := c.Collect(s); err != nil {
		t.Error(err)
	}
	if err := c.Close(); err != nil {
		t.Error(err)
	}
}

type stubCollector struct {
	errid     int
	collected bool
	closed    bool
}

func (c *stubCollector) Collect(*zipkin.Span) error {
	c.collected = true
	if c.errid != 0 {
		return fmt.Errorf("error %d", c.errid)
	}
	return nil
}
func (c *stubCollector) Close() error {
	c.closed = true
	if c.errid != 0 {
		return fmt.Errorf("error %d", c.errid)
	}
	return nil
}

func TestMultiCollector(t *testing.T) {
	cs := zipkin.MultiCollector{
		&stubCollector{errid: 1},
		&stubCollector{},
		&stubCollector{errid: 2},
	}
	err := cs.Collect(s)
	wanted := "error 1; error 2"
	if err == nil || err.Error() != wanted {
		t.Errorf("errors not propagated. got %v, wanted %s", err, wanted)
	}

	for _, c := range cs {
		sc := c.(*stubCollector)
		if !sc.collected {
			t.Error("collect not called")
		}
	}
}

func TestMultiCollectorClose(t *testing.T) {
	cs := zipkin.MultiCollector{
		&stubCollector{errid: 1},
		&stubCollector{},
		&stubCollector{errid: 2},
	}
	err := cs.Close()
	wanted := "error 1; error 2"
	if err == nil || err.Error() != wanted {
		t.Errorf("errors not propagated. got %v, wanted %s", err, wanted)
	}

	for _, c := range cs {
		sc := c.(*stubCollector)
		if !sc.closed {
			t.Error("close not called")
		}
	}
}
