package zipkin

import (
	"errors"
	"strings"
)

// Collector represents a Zipkin trace collector, which is probably a set of
// remote endpoints.
type Collector interface {
	Collect(*Span) error
}

// NopCollector implements Collector but performs no work.
type NopCollector struct{}

// Collect implements Collector.
func (NopCollector) Collect(*Span) error { return nil }

// MultiCollector implements Collector by sending spans to all collectors.
type MultiCollector []Collector

// Collect implements Collector.
func (c MultiCollector) Collect(s *Span) error {
	errs := []string{}
	for _, collector := range c {
		if err := collector.Collect(s); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
