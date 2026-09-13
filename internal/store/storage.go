package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JonasCappe/book-management-api/internal/data"
)

var (
	ErrNotFound                = errors.New("resource not found")
	ErrAuthorNotFound          = errors.New("author not found")
	ErrAuthorsRequired         = errors.New("at least one author is required")
	ErrPublicationDateRequired = errors.New("publication date is required")
)

type Storage struct {
	Books interface {
		Create(context.Context, *data.Book) error
		Update(context.Context, *data.Book) error
		GetByID(context.Context, int64) (*data.Book, error)
		GetAll(context.Context) ([]data.Book, error)
		Delete(context.Context, int64) error
	}
	HistoryEntries interface {
		Create(context.Context, *data.HistoryEntry) error
		GetByID(context.Context, int64) (data.HistoryEntry, error)
		GetByBookID(context.Context, int64) ([]data.HistoryEntry, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Books:          &BookStore{db: db},
		HistoryEntries: &HistoryEntryStore{db: db},
	}
}
