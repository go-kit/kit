package server

// AddService is the abstract representation of this service.
type AddService interface {
	Sum(a, b int) int
	Concat(a, b string) string
}
