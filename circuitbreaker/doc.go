// Package circuitbreaker implements the circuit breaker pattern.
//
// Circuit breakers prevent thundering herds, and improve resiliency against
// intermittent errors. Every client-side endpoint should be wrapped in a
// circuit breaker.
package circuitbreaker
