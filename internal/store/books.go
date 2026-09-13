package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/JonasCappe/book-management-api/internal/data"
)

type BookStore struct {
	db *sql.DB
}

func (s *BookStore) Create(ctx context.Context, newBook *data.Book) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := ensureAuthorExists(ctx, tx, newBook.AuthorID); err != nil {
		return err
	}

	query := `
		INSERT INTO books (title, description, author_id) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at;
	`

	if err := tx.QueryRowContext(ctx, query,
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

	changes, err := json.Marshal(map[string]any{"after": newBook})
	if err != nil {
		return err
	}
	if err := createHistoryEntry(ctx, tx, &data.HistoryEntry{
		BookID:      newBook.ID,
		ChangeType:  data.ChangeTypeCreated,
		Description: "book created",
		Changes:     changes,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *BookStore) Update(ctx context.Context, updatedBook *data.Book) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	previousBook, err := getBookByID(ctx, tx, updatedBook.ID, true)
	if err != nil {
		return err
	}
	if err := ensureAuthorExists(ctx, tx, updatedBook.AuthorID); err != nil {
		return err
	}

	query := `
		UPDATE books
		SET title = $1, description = $2, author_id = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING updated_at;
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		updatedBook.Title,
		updatedBook.Description,
		updatedBook.AuthorID,
		updatedBook.ID,
	).Scan(&updatedBook.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	changes, err := json.Marshal(map[string]any{
		"before": previousBook,
		"after":  updatedBook,
	})
	if err != nil {
		return err
	}

	if err := createHistoryEntry(ctx, tx, &data.HistoryEntry{
		BookID:      updatedBook.ID,
		ChangeType:  data.ChangeTypeUpdated,
		Description: "book updated",
		Changes:     changes,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *BookStore) GetByID(ctx context.Context, id int64) (*data.Book, error) {
	return getBookByID(ctx, s.db, id, false)
}

type bookQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func getBookByID(ctx context.Context, db bookQueryer, id int64, forUpdate bool) (*data.Book, error) {
	query := `
		SELECT id, title, description, author_id, created_at, updated_at
		FROM books
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1;
	`

	book := &data.Book{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.AuthorID,
		&book.CreatedAt,
		&book.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookStore) GetAll(ctx context.Context) ([]data.Book, error) {
	query := `
		SELECT id, title, description, author_id, created_at, updated_at
		FROM books
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := s.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	books := make([]data.Book, 0)

	for rows.Next() {
		var book data.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Description,
			&book.AuthorID,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}

func (s *BookStore) Delete(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	book, err := getBookByID(ctx, tx, id, true)
	if err != nil {
		return err
	}

	changes, err := json.Marshal(map[string]any{"before": book})

	if err != nil {
		return err
	}

	if err := createHistoryEntry(ctx, tx, &data.HistoryEntry{
		BookID:      id,
		ChangeType:  data.ChangeTypeDeleted,
		Description: "book deleted",
		Changes:     changes,
	}); err != nil {
		return err
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE books SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

func ensureAuthorExists(ctx context.Context, db bookQueryer, authorID int64) error {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM authors WHERE id = $1)`, authorID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrAuthorNotFound
	}
	return nil
}
