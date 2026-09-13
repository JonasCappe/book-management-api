package data

import "time"

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	PublicationDate Date      `json:"publication_date"`
	Authors         []Author  `json:"authors"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}


