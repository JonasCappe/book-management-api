package main

import (
	"net/http/httptest"
	"testing"

	"github.com/JonasCappe/book-management-api/internal/data"
)

func TestReadHistoryQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		want      data.HistoryQuery
		wantError bool
	}{
		{
			name:  "defaults",
			query: "",
			want:  data.HistoryQuery{Page: 1, PageSize: 20, Order: "desc"},
		},
		{
			name:  "all options",
			query: "?page=2&page_size=10&change_type=updated&order=asc",
			want: data.HistoryQuery{
				Page:       2,
				PageSize:   10,
				ChangeType: data.ChangeTypeUpdated,
				Order:      "asc",
			},
		},
		{name: "invalid page", query: "?page=0", wantError: true},
		{name: "invalid page size", query: "?page_size=101", wantError: true},
		{name: "invalid change type", query: "?change_type=renamed", wantError: true},
		{name: "invalid order", query: "?order=random", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/books/1/history"+test.query, nil)
			got, err := readHistoryQuery(request)
			if test.wantError {
				if err == nil {
					t.Fatal("readHistoryQuery() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("readHistoryQuery() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("readHistoryQuery() = %#v, want %#v", got, test.want)
			}
		})
	}
}
