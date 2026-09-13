package store

import (
	"context"
	"database/sql"

	"github.com/JonasCappe/book-management-api/internal/data"
)

type Storage struct {
	Books interface {
		Create(context.Context, *data.Book) error
		Update(context.Context, *data.Book) error
		GetByID(context.Context, int64) (data.Book, error)
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
