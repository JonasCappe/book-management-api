package main

import (
	"testing"

	"github.com/JonasCappe/book-management-api/internal/data"
)

func TestNotBlankValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "non-blank value",
			value:   "The Hobbit",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate.Var(tt.value, "notblank")

			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate.Var() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationFields(t *testing.T) {
	payload := CreateBookPayload{
		Title:     "   ",
		AuthorIDs: []int64{1},
	}

	err := Validate.Struct(payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	fields := validationFields(err)

	if got := fields["title"]; got != "must not be blank" {
		t.Errorf(
			"title error = %q, want %q",
			got,
			"must not be blank",
		)
	}

	if got := fields["publication_date"]; got != "is required" {
		t.Errorf(
			"publication_date error = %q, want %q",
			got,
			"is required",
		)
	}
}

func TestValidationFieldsUsesJSONFieldNames(t *testing.T) {
	payload := CreateBookPayload{
		Title:           "The Hobbit",
		PublicationDate: data.Date{ },
		AuthorIDs:       []int64{},
	}

	err := Validate.Struct(payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	fields := validationFields(err)

	if _, ok := fields["author_ids"]; !ok {
		t.Fatalf("expected author_ids field, got %v", fields)
	}
}
