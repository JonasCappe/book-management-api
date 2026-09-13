package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JonasCappe/book-management-api/internal/data"
	"github.com/lib/pq"
)

func TestGetAllRejectsUnsafeSorting(t *testing.T) {
	bookStore := &BookStore{}

	_, err := bookStore.GetAll(context.Background(), data.BookQuery{
		Page: 1, PageSize: 20, OrderBy: "id; DROP TABLE books", Order: "asc",
	})
	if err == nil {
		t.Fatal("GetAll() error = nil, want unsafe sort field rejected")
	}

	_, err = bookStore.GetAll(context.Background(), data.BookQuery{
		Page: 1, PageSize: 20, OrderBy: "id", Order: "asc; DROP TABLE books",
	})
	if err == nil {
		t.Fatal("GetAll() error = nil, want unsafe sort order rejected")
	}
}

func TestDescribeBookChanges(t *testing.T) {
	oldDate, _ := data.ParseDate("1937-09-21")
	newDate, _ := data.ParseDate("1937-09-22")
	before := &data.Book{
		Title:           "The Hobbitt",
		Description:     "Old description",
		PublicationDate: oldDate,
		Authors:         []data.Author{{ID: 1, Name: "Old Author"}},
	}
	after := &data.Book{
		Title:           "The Hobbit",
		Description:     "New description",
		PublicationDate: newDate,
		Authors:         []data.Author{{ID: 2, Name: "J.R.R. Tolkien"}},
	}

	description := describeBookChanges(before, after)
	for _, expected := range []string{
		`Title changed from "The Hobbitt" to "The Hobbit"`,
		`Description changed from "Old description" to "New description"`,
		"Publication date changed from 1937-09-21 to 1937-09-22",
		`Author "J.R.R. Tolkien" was added`,
		`Author "Old Author" was removed`,
	} {
		if !strings.Contains(description, expected) {
			t.Errorf("description %q does not contain %q", description, expected)
		}
	}
}

func TestNormalizeBookWriteError(t *testing.T) {
	databaseError := &pq.Error{Code: "23505", Constraint: "books_title_key"}
	if err := normalizeBookWriteError(databaseError); !errors.Is(err, ErrDuplicateTitle) {
		t.Fatalf("normalizeBookWriteError() = %v, want ErrDuplicateTitle", err)
	}

	otherError := errors.New("database unavailable")
	if err := normalizeBookWriteError(otherError); !errors.Is(err, otherError) {
		t.Fatalf("normalizeBookWriteError() = %v, want original error", err)
	}
}
