package catalog

import (

	"errors"
	"github.com/go-kit/kit/examples/library/books"
)

var ErrInvalidArgument = errors.New("invalid argument")



type Service interface {

	AddBook(book books.Book) (bool, error)
	GetBook(name string) ([]*books.Book, error)
}

type service struct {
	books         books.Repository
}

func (s *service) AddBook(book books.Book) (bool, error) {

	return s.books.AddBook(book)
}


func (s *service) GetBook(name string) ([]*books.Book, error) {

	return s.books.GetBook(name)
}

func NewService(br books.Repository) Service {

	return  &service{books: br}
}



