package data

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}


type HistoryQuery struct {
	Page       int 			`json:"offset" validate:"gte=0"`
	PageSize   int			`json:"limit" validate:"gte=1,lte=20"`
	ChangeType ChangeType
	Order      string		`json:"sort" validate:"oneof=asc desc"`
}


type BookQuery struct {
	Page     int
	PageSize int

	Title  string
	Author string

	OrderBy string
	Order   string
}


type HistoryPage struct {
	Entries    []HistoryEntry `json:"entries"`
	Pagination Pagination     `json:"pagination"`
}

type BookPage struct {
	Entries []Book `json:"entries"`
	Pagination Pagination     `json:"pagination"`
}
