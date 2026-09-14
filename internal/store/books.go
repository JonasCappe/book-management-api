package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/JonasCappe/book-management-api/internal/data"
	"github.com/lib/pq"
)

type BookStore struct {
	db *sql.DB
}

type bookQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *BookStore) Create(ctx context.Context, book *data.Book) error {
	if book.PublicationDate.IsZero() {
		return ErrPublicationDateRequired
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	authorIDs := authorIDs(book.Authors)
	if err := ensureAuthorsExist(ctx, tx, authorIDs); err != nil {
		return err
	}

	query := `
		INSERT INTO books (title, description, publication_date)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at;
	`
	if err := tx.QueryRowContext(
		ctx,
		query,
		book.Title,
		book.Description,
		book.PublicationDate,
	).Scan(&book.ID, &book.CreatedAt, &book.UpdatedAt); err != nil {
		return normalizeBookWriteError(err)
	}

	if err := replaceBookAuthors(ctx, tx, book.ID, authorIDs); err != nil {
		return err
	}
	book.Authors, err = getAuthorsForBook(ctx, tx, book.ID)
	if err != nil {
		return normalizeBookWriteError(err)
	}

	description := fmt.Sprintf("Book %q was created", book.Title)
	if err := recordBookHistory(ctx, tx, book.ID, data.ChangeTypeCreated, description, map[string]any{"after": book}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BookStore) Update(ctx context.Context, book *data.Book) error {
	if book.PublicationDate.IsZero() {
		return ErrPublicationDateRequired
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	previousBook, err := getBookByID(ctx, tx, book.ID, true)
	if err != nil {
		return err
	}

	authorIDs := authorIDs(book.Authors)
	if err := ensureAuthorsExist(ctx, tx, authorIDs); err != nil {
		return err
	}

	query := `
		UPDATE books
		SET title = $1, description = $2, publication_date = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING updated_at;
	`
	err = tx.QueryRowContext(
		ctx,
		query,
		book.Title,
		book.Description,
		book.PublicationDate,
		book.ID,
	).Scan(&book.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return normalizeBookWriteError(err)
	}

	if err := replaceBookAuthors(ctx, tx, book.ID, authorIDs); err != nil {
		return err
	}
	book.Authors, err = getAuthorsForBook(ctx, tx, book.ID)
	if err != nil {
		return err
	}

	if err := recordBookHistory(ctx, tx, book.ID, data.ChangeTypeUpdated, describeBookChanges(previousBook, book), map[string]any{
		"before": previousBook,
		"after":  book,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *BookStore) GetByID(ctx context.Context, id int64) (*data.Book, error) {
	return getBookByID(ctx, s.db, id, false)
}

func getBookByID(ctx context.Context, db bookQueryer, id int64, forUpdate bool) (*data.Book, error) {
	query := `
		SELECT id, title, description, publication_date, created_at, updated_at
		FROM books
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1;
	`
	if forUpdate {
		query = `
			SELECT id, title, description, publication_date, created_at, updated_at
			FROM books
			WHERE id = $1 AND deleted_at IS NULL
			FOR UPDATE;
		`
	}

	book := &data.Book{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.PublicationDate,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	book.Authors, err = getAuthorsForBook(ctx, db, book.ID)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (s *BookStore) GetAll(ctx context.Context, filters data.BookQuery) (data.BookPage, error) {
	sortColumns := map[string]string{
		"id":               "id",
		"title":            "title",
		"publication_date": "publication_date",
		"created_at":       "created_at",
		"updated_at":       "updated_at",
	}
	orderBy, valid := sortColumns[filters.OrderBy]
	if !valid {
		return data.BookPage{}, fmt.Errorf("unsupported book sort field %q", filters.OrderBy)
	}
	order := strings.ToUpper(filters.Order)
	if order != "ASC" && order != "DESC" {
		return data.BookPage{}, fmt.Errorf("unsupported book sort order %q", filters.Order)
	}

	conditions := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 6)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if filters.Title != "" {
		addCondition("title ILIKE '%%' || $%d || '%%'", filters.Title)
	}
	if filters.AuthorID != 0 {
		addCondition("EXISTS (SELECT 1 FROM book_authors ba WHERE ba.book_id = books.id AND ba.author_id = $%d)", filters.AuthorID)
	}
	if filters.PublishedFrom != nil {
		addCondition("publication_date >= $%d", *filters.PublishedFrom)
	}
	if filters.PublishedTo != nil {
		addCondition("publication_date <= $%d", *filters.PublishedTo)
	}

	whereClause := strings.Join(conditions, " AND ")
	var totalItems int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books WHERE "+whereClause, args...).Scan(&totalItems); err != nil {
		return data.BookPage{}, err
	}

	offset := (filters.Page - 1) * filters.PageSize
	args = append(args, filters.PageSize, offset)
	secondaryOrder := ""
	if orderBy != "id" {
		secondaryOrder = ", id " + order
	}
	query := fmt.Sprintf(`
		SELECT id, title, description, publication_date, created_at, updated_at
		FROM books
		WHERE %s
		ORDER BY %s %s%s
		LIMIT $%d OFFSET $%d;
	`, whereClause, orderBy, order, secondaryOrder, len(args)-1, len(args))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return data.BookPage{}, err
	}
	defer rows.Close()

	books := make([]data.Book, 0)
	for rows.Next() {
		var book data.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Description,
			&book.PublicationDate,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return data.BookPage{}, err
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return data.BookPage{}, err
	}

	for i := range books {
		books[i].Authors, err = getAuthorsForBook(ctx, s.db, books[i].ID)
		if err != nil {
			return data.BookPage{}, err
		}
	}
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + filters.PageSize - 1) / filters.PageSize
	}
	return data.BookPage{
		Books: books,
		Pagination: data.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
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
	description := fmt.Sprintf("Book %q was deleted", book.Title)
	if err := recordBookHistory(ctx, tx, id, data.ChangeTypeDeleted, description, map[string]any{"before": book}); err != nil {
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

func ensureAuthorsExist(ctx context.Context, db bookQueryer, authorIDs []int64) error {
	if len(authorIDs) == 0 {
		return ErrAuthorsRequired
	}

	var count int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM authors WHERE id = ANY($1)`,
		pq.Array(authorIDs),
	).Scan(&count); err != nil {
		return err
	}
	if count != len(authorIDs) {
		return ErrAuthorNotFound
	}
	return nil
}

func authorIDs(authors []data.Author) []int64 {
	seen := make(map[int64]struct{}, len(authors))
	ids := make([]int64, 0, len(authors))
	for _, author := range authors {
		if _, exists := seen[author.ID]; exists {
			continue
		}
		seen[author.ID] = struct{}{}
		ids = append(ids, author.ID)
	}
	return ids
}

func replaceBookAuthors(ctx context.Context, tx *sql.Tx, bookID int64, authorIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM book_authors WHERE book_id = $1`, bookID); err != nil {
		return err
	}
	for _, authorID := range authorIDs {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`,
			bookID,
			authorID,
		); err != nil {
			return err
		}
	}
	return nil
}

func getAuthorsForBook(ctx context.Context, db bookQueryer, bookID int64) ([]data.Author, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT authors.id, authors.name, authors.bio
		FROM authors
		INNER JOIN book_authors ON book_authors.author_id = authors.id
		WHERE book_authors.book_id = $1
		ORDER BY authors.id;
	`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors := make([]data.Author, 0)
	for rows.Next() {
		var author data.Author
		if err := rows.Scan(&author.ID, &author.Name, &author.Bio); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return authors, nil
}

func recordBookHistory(
	ctx context.Context,
	tx *sql.Tx,
	bookID int64,
	changeType data.ChangeType,
	description string,
	changesValue any,
) error {
	changes, err := json.Marshal(changesValue)
	if err != nil {
		return err
	}
	return createHistoryEntry(ctx, tx, &data.HistoryEntry{
		BookID:      bookID,
		ChangeType:  changeType,
		Description: description,
		Changes:     changes,
	})
}

func normalizeBookWriteError(err error) error {
	var pqError *pq.Error
	if errors.As(err, &pqError) && pqError.Code == "23505" && pqError.Constraint == "books_title_key" {
		return ErrDuplicateTitle
	}
	return err
}

func describeBookChanges(before, after *data.Book) string {
	descriptions := make([]string, 0)
	if before.Title != after.Title {
		descriptions = append(descriptions, fmt.Sprintf("Title changed from %q to %q", before.Title, after.Title))
	}
	if before.Description != after.Description {
		descriptions = append(descriptions, fmt.Sprintf("Description changed from %q to %q", before.Description, after.Description))
	}
	if !before.PublicationDate.Equal(after.PublicationDate.Time) {
		descriptions = append(descriptions, fmt.Sprintf(
			"Publication date changed from %s to %s",
			before.PublicationDate.Format("2006-01-02"),
			after.PublicationDate.Format("2006-01-02"),
		))
	}

	beforeAuthors := make(map[int64]data.Author, len(before.Authors))
	afterAuthors := make(map[int64]data.Author, len(after.Authors))
	for _, author := range before.Authors {
		beforeAuthors[author.ID] = author
	}
	for _, author := range after.Authors {
		afterAuthors[author.ID] = author
		if _, existed := beforeAuthors[author.ID]; !existed {
			descriptions = append(descriptions, fmt.Sprintf("Author %q was added", author.Name))
		}
	}
	for _, author := range before.Authors {
		if _, remains := afterAuthors[author.ID]; !remains {
			descriptions = append(descriptions, fmt.Sprintf("Author %q was removed", author.Name))
		}
	}

	if len(descriptions) == 0 {
		return "Book was updated without visible field changes"
	}
	return strings.Join(descriptions, "; ")
}
