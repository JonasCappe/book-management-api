package main

import (
	"testing"

	"github.com/JonasCappe/book-management-api/internal/data"
	"github.com/go-playground/validator/v10"
)

func newTestValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	configureValidator(validate)
	return validate
}

func TestCreateBookPayloadValidation(t *testing.T) {
	publicationDate, err := data.ParseDate("1937-09-21")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		payload CreateBookPayload
		valid   bool
	}{
		{
			name: "valid",
			payload: CreateBookPayload{
				Title:           "The Hobbit",
				Description:     "A fantasy novel",
				PublicationDate: publicationDate,
				AuthorIDs:       []int64{1, 2},
			},
			valid: true,
		},
		{name: "blank title", payload: CreateBookPayload{Title: "   ", PublicationDate: publicationDate, AuthorIDs: []int64{1}}},
		{name: "missing publication date", payload: CreateBookPayload{Title: "The Hobbit", AuthorIDs: []int64{1}}},
		{name: "missing authors", payload: CreateBookPayload{Title: "The Hobbit", PublicationDate: publicationDate}},
		{name: "duplicate authors", payload: CreateBookPayload{Title: "The Hobbit", PublicationDate: publicationDate, AuthorIDs: []int64{1, 1}}},
		{name: "invalid author ID", payload: CreateBookPayload{Title: "The Hobbit", PublicationDate: publicationDate, AuthorIDs: []int64{0}}},
	}

	validate := newTestValidator()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validate.Struct(test.payload)
			if test.valid && err != nil {
				t.Fatalf("validation error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("validation error = nil, want error")
			}
		})
	}
}

func TestUpdateBookPayloadValidation(t *testing.T) {
	emptyTitle := "  "
	emptyAuthors := []int64{}
	validate := newTestValidator()

	if err := validate.Struct(UpdateBookPayload{Title: &emptyTitle}); err == nil {
		t.Fatal("blank title validation error = nil, want error")
	}
	if err := validate.Struct(UpdateBookPayload{AuthorIDs: &emptyAuthors}); err == nil {
		t.Fatal("empty authors validation error = nil, want error")
	}
}
