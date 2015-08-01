package publisher

import "github.com/go-kit/kit/endpoint"

// Publisher produces endpoints.
type Publisher interface {
	Subscribe(chan<- []endpoint.Endpoint)
	Unsubscribe(chan<- []endpoint.Endpoint)
	Stop()
}
