// Package routing provides the routing domain service. It does not actually
// implement the routing service but merely acts as a proxy for a separate
// bounded context.
package routing

import (
	"time"

	"github.com/marcusolsson/goddd/cargo"
)

// Service provides access to an external routing service.
type Service interface {
	// FetchRoutesForSpecification finds all possible routes that satisfy a
	// given specification.
	FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary
}

// Route is a read model for routing views.
type Route struct {
	Legs []Leg `json:"legs"`
}

// Leg is a read model for routing views.
type Leg struct {
	VoyageNumber string    `json:"voyage_number"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	LoadTime     time.Time `json:"load_time"`
	UnloadTime   time.Time `json:"unload_time"`
}
