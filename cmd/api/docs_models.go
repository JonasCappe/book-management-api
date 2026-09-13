package main

import "github.com/JonasCappe/book-management-api/internal/data"

type HealthData struct {
	Status  string `json:"status" example:"ok"`
	Env     string `json:"env" example:"development"`
	Version string `json:"version" example:"0.0.1"`
}

type HealthResponse struct {
	Data HealthData `json:"data"`
}

type BookResponse struct {
	Data data.Book `json:"data"`
}

type BooksResponse struct {
	Data data.BookPage `json:"data"`
}

type HistoryResponse struct {
	Data data.HistoryPage `json:"data"`
}
