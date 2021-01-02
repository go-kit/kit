package books

import (
	"errors"
	"strings"
	"github.com/pborman/uuid"
)

// TrackingID uniquely identifies a particular cargo.
type TrackingID string

// Cargo is the central class in the domain model.
type Book struct {
	Id string
	Name string
	Author string
}


type Repository interface {
	AddBook(book Book) (bool, error)
	GetBook(name string) ([]*Book, error)
	//FindAll() []*Book
}

// ErrUnknown is used when a cargo could not be found.
var ErrUnknown = errors.New("unknown book")

func NextTrackingID() TrackingID {
	return TrackingID(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}
