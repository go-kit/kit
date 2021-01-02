package inmemory

import (
	"sync"
	"github.com/go-kit/kit/examples/library/books"
)

type bookRepository struct {
	mtx    sync.RWMutex
	booksMap map[string]*books.Book
}


func (r *bookRepository) AddBook(book books.Book) (bool, error) {

	r.booksMap[book.Name] = &book
	return true, nil
}


func (r *bookRepository) GetBook(name string) ([]*books.Book, error) {
	bookList := []*books.Book{}
	if len(name) > 0 {
		bookList = append(bookList,r.booksMap[name] )
	} else{

		for _, v :=range r.booksMap {
			bookList = append(bookList, v)
		}
	}
	return bookList, nil
}

// NewCargoRepository returns a new instance of a in-memory cargo repository.
func NewBookRepository() books.Repository {
	return &bookRepository{
		booksMap: make(map[string]*books.Book),
	}
}
