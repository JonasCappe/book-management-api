package data

import (
	"encoding/json"
	"time"
)

type ChangeType string

const (
	ChangeTypeCreated ChangeType = "created"
	ChangeTypeUpdated ChangeType = "updated"
	ChangeTypeDeleted ChangeType = "deleted"
)

type HistoryEntry struct {
	ID          int64           `json:"id"`
	BookID      int64           `json:"book_id"`
	ChangedAt   time.Time       `json:"changed_at"`
	ChangeType  ChangeType      `json:"change_type"`
	Description string          `json:"description"`
	Changes     json.RawMessage `json:"changes"`
}

type HistoryQuery struct {
	Page       int
	PageSize   int
	ChangeType ChangeType
	Order      string
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type HistoryPage struct {
	Entries    []HistoryEntry `json:"entries"`
	Pagination Pagination     `json:"pagination"`
}
