package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

func (s *HistoryEntryStore) GetByBookID(ctx context.Context, bookID int64, filters data.HistoryQuery) (data.HistoryPage, error) {
	var bookExists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM books WHERE id = $1)`, bookID).Scan(&bookExists); err != nil {
		return data.HistoryPage{}, err
	}
	if !bookExists {
		return data.HistoryPage{}, ErrNotFound
	}

	where := "book_id = $1"
	args := []any{bookID}
	if filters.ChangeType != "" {
		where += " AND change_type = $2"
		args = append(args, filters.ChangeType)
	}

	var totalItems int
	countQuery := "SELECT COUNT(*) FROM history_entries WHERE " + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return data.HistoryPage{}, err
	}

	order := "DESC"
	if filters.Order == "asc" {
		order = "ASC"
	}
	query := fmt.Sprintf(`
		SELECT
			id,
			book_id,
			changed_at,
			change_type,
			description,
			changes
		FROM history_entries
		WHERE %s
		ORDER BY changed_at %s, id %s
		LIMIT $%d OFFSET $%d;
	`, where, order, order, len(args)+1, len(args)+2)
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {
		return data.HistoryPage{}, err
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
			return data.HistoryPage{}, err
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return data.HistoryPage{}, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + filters.PageSize - 1) / filters.PageSize
	}
	return data.HistoryPage{
		Entries: entries,
		Pagination: data.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
}
