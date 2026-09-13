package main

import (
	"errors"
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
		app.badRequestError(w, r, err)
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

	book, err := app.store.Books.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
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
		app.badRequestError(w, r, err)
		return
	}

	if payload.Title != nil {
		book.Title = *payload.Title
	}
	if payload.Description != nil {
		book.Description = *payload.Description
	}
	if payload.PublicationDate != nil {
		book.PublicationDate = *payload.PublicationDate
	}
	if payload.AuthorIDs != nil {
		book.Authors = authorsFromIDs(*payload.AuthorIDs)
	}

	if err := app.store.Books.Update(r.Context(), book); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		case errors.Is(err, store.ErrDuplicateTitle):
			app.conflictError(w, r, err)
		case errors.Is(err, store.ErrAuthorNotFound),
			errors.Is(err, store.ErrAuthorsRequired):
			app.validationError(w, r, map[string]string{"author_ids": err.Error()})
		case errors.Is(err, store.ErrPublicationDateRequired):
			app.validationError(w, r, map[string]string{"publication_date": err.Error()})
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
//	@Success		200	{object}	BooksResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/v1/books [get]
func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := app.store.Books.GetAll(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, books); err != nil {
		app.internalServerError(w, r, err)
	}
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
