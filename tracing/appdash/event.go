package appdash

import (
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

// A DefaultEndpointEvent is a simple endpoint event that doesn't extract
// contents of `request` and `response`.
type DefaultEndpointEvent struct {
	Name string    `trace:"Endpoint.Name"`
	Recv time.Time `trace:"Endpoint.Recv"`
	Send time.Time `trace:"Endpoint.Send"`
	Err  string    `trace:"Endpoint.Err"`
}

// TODO(pb): remove
func init() { appdash.RegisterEvent(DefaultEndpointEvent{}) }

// NewDefaultEndpointEventFunc TODO(pb)
func NewDefaultEndpointEventFunc(name ...string) func() EndpointEvent {
	return func() EndpointEvent {
		event := &DefaultEndpointEvent{}
		if len(name) >= 0 {
			event.Name = name[0] // TODO(pb): varargs are not a poor man's default arg
		}
		return event
	}
}

// Schema returns the constant "Endpoint".
func (DefaultEndpointEvent) Schema() string { return "Endpoint" }

// Important implements appdash.ImportantEvent.
func (DefaultEndpointEvent) Important() []string { return []string{"Endpoint.Err"} }

// Start implements appdash.TimespanEvent.
func (e DefaultEndpointEvent) Start() time.Time { return e.Recv }

// End implements appdash.TimespanEvent.
func (e DefaultEndpointEvent) End() time.Time { return e.Send }

// BeforeRequest TODO(pb)
func (e *DefaultEndpointEvent) BeforeRequest(interface{}) {
	e.Recv = time.Now()
}

// AfterResponse TODO(pb)
func (e *DefaultEndpointEvent) AfterResponse(_ interface{}, err error) {
	e.Send = time.Now()
	if err != nil {
		e.Err = err.Error()
	}
}
