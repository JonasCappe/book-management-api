package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JonasCappe/book-management-api/internal/data"
	"github.com/JonasCappe/book-management-api/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreateBookPayload struct {
	Title           string    `json:"title" validate:"required,notblank,max=150"`
	Description     string    `json:"description" validate:"max=1024"`
	PublicationDate data.Date `json:"publication_date" validate:"required" swaggertype:"string" example:"1937-09-21"`
	AuthorIDs       []int64   `json:"author_ids" validate:"required,min=1,unique,dive,gt=0"`
}

type UpdateBookPayload struct {
	Title           *string    `json:"title" validate:"omitempty,notblank,max=150"`
	Description     *string    `json:"description" validate:"omitempty,max=1024"`
	PublicationDate *data.Date `json:"publication_date" validate:"omitempty" swaggertype:"string" example:"1937-09-21"`
	AuthorIDs       *[]int64   `json:"author_ids" validate:"omitempty,min=1,unique,dive,gt=0"`
}

// createBookHandler godoc
//
//	@Summary		Create a book
//	@Description	Creates a book with a publication date and one or more existing authors.
//	@Tags			books
//	@Accept			json
//	@Produce		json
//	@Param			book	body		CreateBookPayload	true	"Book to create"
//	@Success		201		{object}	BookResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		422		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/books [post]
func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateBookPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.validationError(w, r, validationFields(err))
		return
	}

	book := &data.Book{
		Title:           payload.Title,
		Description:     payload.Description,
		PublicationDate: payload.PublicationDate,
		Authors:         authorsFromIDs(payload.AuthorIDs),
	}

	ctx := r.Context()

	if err := app.store.Books.Create(ctx, book); err != nil {
		if errors.Is(err, store.ErrDuplicateTitle) {
			app.conflictError(w, r, err)
			return
		}
		if errors.Is(err, store.ErrAuthorNotFound) ||
			errors.Is(err, store.ErrAuthorsRequired) {
			app.validationError(w, r, map[string]string{"author_ids": err.Error()})
			return
		}
		if errors.Is(err, store.ErrPublicationDateRequired) {
			app.validationError(w, r, map[string]string{"publication_date": err.Error()})
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	w.Header().Set("Location", "/v1/books/"+strconv.FormatInt(book.ID, 10))

	if err := app.jsonResponse(w, http.StatusCreated, book); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getBookHandler godoc
//
//	@Summary		Retrieve a book
//	@Description	Returns an active book and its authors.
//	@Tags			books
//	@Produce		json
//	@Param			bookID	path		int	true	"Book ID"
//	@Success		200		{object}	BookResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/books/{bookID} [get]
func (app *application) getBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readBookID(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	book, err := app.store.Books.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, book); err != nil {
		app.internalServerError(w, r, err)
	}
}

// deleteBookHandler godoc
//
//	@Summary		Delete a book
//	@Description	Soft-deletes a book while preserving its change history.
//	@Tags			books
//	@Param			bookID	path	int	true	"Book ID"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/v1/books/{bookID} [delete]
func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readBookID(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Books.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// patchBookHandler godoc
//
//	@Summary		Update a book
//	@Description	Partially updates a book and records a human-readable history entry.
//	@Tags			books
//	@Accept			json
//	@Produce		json
//	@Param			bookID	path		int					true	"Book ID"
//	@Param			book	body		UpdateBookPayload	true	"Fields to update"
//	@Success		200		{object}	BookResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		422		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/books/{bookID} [patch]
func (app *application) patchBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readBookID(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	var payload UpdateBookPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if payload.Title == nil && payload.Description == nil && payload.PublicationDate == nil && payload.AuthorIDs == nil {
		app.validationError(w, r, map[string]string{"body": "must contain at least one field"})
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.validationError(w, r, validationFields(err))
		return
	}

	update := data.BookUpdate{
		Title:           payload.Title,
		Description:     payload.Description,
		PublicationDate: payload.PublicationDate,
		AuthorIDs:       payload.AuthorIDs,
	}

	book, err := app.store.Books.Update(r.Context(), id, update)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)

		case errors.Is(err, store.ErrDuplicateTitle):
			app.conflictError(w, r, err)

		case errors.Is(err, store.ErrAuthorNotFound),
			errors.Is(err, store.ErrAuthorsRequired):
			app.validationError(
				w,
				r,
				map[string]string{"author_ids": err.Error()},
			)

		case errors.Is(err, store.ErrPublicationDateRequired):
			app.validationError(
				w,
				r,
				map[string]string{"publication_date": err.Error()},
			)

		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	if err := app.jsonResponse(w, http.StatusOK, book); err != nil {
		app.internalServerError(w, r, err)
	}
}

// getBooksHandler godoc
//
//	@Summary		List books
//	@Description	Returns all active books and their authors.
//	@Tags			books
//	@Produce		json
//	@Param			page			query		int		false	"Page number"		default(1)	minimum(1)
//	@Param			page_size		query		int		false	"Items per page"	default(20)	minimum(1)	maximum(100)
//	@Param			title			query		string	false	"Partial title match"
//	@Param			author_id		query		int		false	"Author ID"					minimum(1)
//	@Param			published_from	query		string	false	"Published on or after"		Format(date)
//	@Param			published_to	query		string	false	"Published on or before"	Format(date)
//	@Param			order_by		query		string	false	"Sort field"				default(id)		Enums(id, title, publication_date, created_at, updated_at)
//	@Param			order			query		string	false	"Sort order"				default(asc)	Enums(asc, desc)
//	@Success		200				{object}	BooksResponse
//	@Failure		400				{object}	ErrorResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/v1/books [get]
func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	filters, err := readBookQuery(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	page, err := app.store.Books.GetAll(r.Context(), filters)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, page); err != nil {
		app.internalServerError(w, r, err)
	}
}

func readBookQuery(r *http.Request) (data.BookQuery, error) {
	query := data.BookQuery{Page: 1, PageSize: 20, OrderBy: "id", Order: "asc"}
	values := r.URL.Query()

	if value := values.Get("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil || page < 1 {
			return query, errors.New("page must be a positive integer")
		}
		query.Page = page
	}
	if value := values.Get("page_size"); value != "" {
		pageSize, err := strconv.Atoi(value)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return query, errors.New("page_size must be between 1 and 100")
		}
		query.PageSize = pageSize
	}
	query.Title = values.Get("title")
	if value := values.Get("author_id"); value != "" {
		authorID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || authorID < 1 {
			return query, errors.New("author_id must be a positive integer")
		}
		query.AuthorID = authorID
	}
	if value := values.Get("published_from"); value != "" {
		publishedFrom, err := data.ParseDate(value)
		if err != nil {
			return query, fmt.Errorf("published_from %w", err)
		}
		query.PublishedFrom = &publishedFrom
	}
	if value := values.Get("published_to"); value != "" {
		publishedTo, err := data.ParseDate(value)
		if err != nil {
			return query, fmt.Errorf("published_to %w", err)
		}
		query.PublishedTo = &publishedTo
	}
	if query.PublishedFrom != nil && query.PublishedTo != nil && query.PublishedFrom.After(query.PublishedTo.Time) {
		return query, errors.New("published_from must be on or before published_to")
	}
	if value := values.Get("order_by"); value != "" {
		switch value {
		case "id", "title", "publication_date", "created_at", "updated_at":
			query.OrderBy = value
		default:
			return query, errors.New("order_by must be id, title, publication_date, created_at, or updated_at")
		}
	}
	if value := values.Get("order"); value != "" {
		if value != "asc" && value != "desc" {
			return query, errors.New("order must be asc or desc")
		}
		query.Order = value
	}
	return query, nil
}

func (app *application) readBookID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "bookID"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("book ID must be a positive integer")
	}
	return id, nil
}

func authorsFromIDs(authorIDs []int64) []data.Author {
	authors := make([]data.Author, len(authorIDs))
	for i, id := range authorIDs {
		authors[i] = data.Author{ID: id}
	}
	return authors
}
