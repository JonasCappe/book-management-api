package data

import "time"

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	PublicationDate Date      `json:"publication_date" swaggertype:"string" example:"1937-09-21"`
	Authors         []Author  `json:"authors"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
