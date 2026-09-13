package store

import (
	"context"
	"database/sql"

	"github.com/JonasCappe/book-management-api/internal/data"
)

type BookStore struct {
	db *sql.DB
}

func (s *BookStore) Create(ctx context.Context, newBook *data.Book) (err error) {
	query := `
		INSERT INTO books (title, description, author_id) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at;
	`

	if err := s.db.QueryRowContext(ctx, query,
		newBook.Title,
		newBook.Description,
		newBook.AuthorID,
	).Scan(
		&newBook.ID,
		&newBook.CreatedAt,
		&newBook.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *BookStore) Update(ctx context.Context, updatedBook *data.Book) error {
	query := `
		UPDATE books
		SET title = $1, description = $2, author_id = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at;
	`

	return s.db.QueryRowContext(
		ctx,
		query,
		updatedBook.Title,
		updatedBook.Description,
		updatedBook.AuthorID,
		updatedBook.ID,
	).Scan(&updatedBook.UpdatedAt)
}

func (s *BookStore) GetByID(ctx context.Context, id int64) (book data.Book, err error) {
	query := `
		SELECT id, title, description, author_id, created_at, updated_at
		FROM books
		WHERE id = $1 LIMIT 1;
	`

	err = s.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.AuthorID,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	return book, err
}
