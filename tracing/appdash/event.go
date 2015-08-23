package appdash

import (
	"sourcegraph.com/sourcegraph/appdash"
	"time"
)

// A DefaultEndpointEvent is a simple endpoint event that doesn't extract contents of `request` and `response`.
type DefaultEndpointEvent struct {
	Name string    `trace:"Endpoint.Name"`
	Recv time.Time `trace:"Endpoint.Recv"`
	Send time.Time `trace:"Endpoint.Send"`
	Err  string    `trace:"Endpoint.Err"`
}

func init() { appdash.RegisterEvent(DefaultEndpointEvent{}) }

func NewDefaultEndpointEventFunc(name ...string) func() EndpointEvent {
	return func() EndpointEvent {
		event := &DefaultEndpointEvent{}
		if len(name) >= 0 {
			event.Name = name[0]
		}
		return event
	}
}

// Schema returns the constant "Endpoint"
func (DefaultEndpointEvent) Schema() string { return "Endpoint" }

// Important implements the appdash ImportantEvent
func (DefaultEndpointEvent) Important() []string {
	return []string{"Endpoint.Err"}
}

// Start implements the appdash TimespanEvent interface.
func (e DefaultEndpointEvent) Start() time.Time { return e.Recv }

// End implements the appdash TimespanEvent interface.
func (e DefaultEndpointEvent) End() time.Time { return e.Send }

func (e *DefaultEndpointEvent) BeforeRequest(_ interface{}) {
	e.Recv = time.Now()
}

func (e *DefaultEndpointEvent) AfterResponse(_ interface{}, err error) {
	e.Send = time.Now()
	if err != nil {
		e.Err = err.Error()
	}
}
