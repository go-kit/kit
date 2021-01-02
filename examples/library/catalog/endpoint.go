package catalog

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/library/books"

)

type bookRequest struct {
	Name string `json:"Name,omitempty"`
	Author string `json:"Author,omitempty"`
}

type bookResponse struct {
	Id  string `json:"Id"`
	Added bool `json:"Added,omitempty"`
	Err error  `json:"error,omitempty"`
}

func (r bookResponse) error() error { return r.Err }

func makeBookEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookRequest)
		_, err := s.AddBook(books.Book{Name: req.Name,Author: req.Author})
		return bookResponse{Added:true,Err: err}, nil
	}
}

type getBookRequest struct {}

func makeGetBookEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(bookRequest)
		book, err := s.GetBook(req.Name)
		return book, err

	}
}
