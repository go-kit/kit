package repository

import "github.com/go-kit/kit/examples/shipping/domain/model/cargo"

// CargoRepository provides access a cargo store.
type CargoRepository interface {
	Store(cargo *cargo.Cargo) error
	Find(id cargo.TrackingID) (*cargo.Cargo, error)
	FindAll() []*cargo.Cargo
}
