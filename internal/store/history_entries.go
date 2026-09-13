package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/JonasCappe/book-management-api/internal/data"
)

type HistoryEntryStore struct {
	db *sql.DB
}

type historyEntryCreator interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

var ErrInvalidHistoryChanges = errors.New("history changes must be valid JSON")

func (s *HistoryEntryStore) Create(ctx context.Context, entry *data.HistoryEntry) error {
	return createHistoryEntry(ctx, s.db, entry)
}

func createHistoryEntry(ctx context.Context, db historyEntryCreator, entry *data.HistoryEntry) error {
	if !json.Valid(entry.Changes) {
		return ErrInvalidHistoryChanges
	}

	query := `
		INSERT INTO history_entries (
			book_id,
			change_type,
			description,
			changes
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, changed_at;
	`

	return db.QueryRowContext(
		ctx,
		query,
		entry.BookID,
		entry.ChangeType,
		entry.Description,
		entry.Changes,
	).Scan(
		&entry.ID,
		&entry.ChangedAt,
	)
}

func (s *HistoryEntryStore) GetByID(ctx context.Context, id int64) (data.HistoryEntry, error) {
	query := `
		SELECT
			id,
			book_id,
			changed_at,
			change_type,
			description,
			changes
		FROM history_entries
		WHERE id = $1;
	`

	var entry data.HistoryEntry

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.BookID,
		&entry.ChangedAt,
		&entry.ChangeType,
		&entry.Description,
		&entry.Changes,
	)

	return entry, err
}

func (s *HistoryEntryStore) GetByBookID(ctx context.Context, bookID int64) ([]data.HistoryEntry, error) {
	query := `
		SELECT
			id,
			book_id,
			changed_at,
			change_type,
			description,
			changes
		FROM history_entries
		WHERE book_id = $1
		ORDER BY changed_at DESC, id DESC;
	`

	rows, err := s.db.QueryContext(ctx, query, bookID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]data.HistoryEntry, 0)

	for rows.Next() {
		var entry data.HistoryEntry

		err := rows.Scan(
			&entry.ID,
			&entry.BookID,
			&entry.ChangedAt,
			&entry.ChangeType,
			&entry.Description,
			&entry.Changes,
		)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
