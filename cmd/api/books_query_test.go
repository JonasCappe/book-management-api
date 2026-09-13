package main

import (
	"net/http/httptest"
	"testing"

	"github.com/JonasCappe/book-management-api/internal/data"
)

func TestReadBookQuery(t *testing.T) {
	from, _ := data.ParseDate("1930-01-01")
	to, _ := data.ParseDate("1940-12-31")
	tests := []struct {
		name      string
		query     string
		want      data.BookQuery
		wantError bool
	}{
		{
			name: "defaults",
			want: data.BookQuery{Page: 1, PageSize: 20, OrderBy: "id", Order: "asc"},
		},
		{
			name:  "all options",
			query: "?page=2&page_size=10&title=ring&author_id=3&published_from=1930-01-01&published_to=1940-12-31&order_by=publication_date&order=desc",
			want: data.BookQuery{
				Page: 2, PageSize: 10, Title: "ring", AuthorID: 3,
				PublishedFrom: &from, PublishedTo: &to,
				OrderBy: "publication_date", Order: "desc",
			},
		},
		{name: "invalid page", query: "?page=0", wantError: true},
		{name: "invalid page size", query: "?page_size=101", wantError: true},
		{name: "invalid author", query: "?author_id=nope", wantError: true},
		{name: "invalid from date", query: "?published_from=01-01-1930", wantError: true},
		{name: "invalid to date", query: "?published_to=tomorrow", wantError: true},
		{name: "reversed date range", query: "?published_from=1940-01-01&published_to=1930-01-01", wantError: true},
		{name: "invalid sort field", query: "?order_by=deleted_at", wantError: true},
		{name: "invalid order", query: "?order=random", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/books"+test.query, nil)
			got, err := readBookQuery(request)
			if test.wantError {
				if err == nil {
					t.Fatal("readBookQuery() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("readBookQuery() error = %v", err)
			}
			if got.Page != test.want.Page || got.PageSize != test.want.PageSize || got.Title != test.want.Title ||
				got.AuthorID != test.want.AuthorID || got.OrderBy != test.want.OrderBy || got.Order != test.want.Order ||
				!sameDate(got.PublishedFrom, test.want.PublishedFrom) || !sameDate(got.PublishedTo, test.want.PublishedTo) {
				t.Fatalf("readBookQuery() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func sameDate(left, right *data.Date) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(right.Time)
}
