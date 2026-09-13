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
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorID    int64  `json:"author_id"`
}

type UpdateBookPayload struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	AuthorID    *int64  `json:"author_id"`
}

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateBookPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	book := &data.Book{
		Title:       payload.Title,
		Description: payload.Description,
		AuthorID:    payload.AuthorID,
	}

	ctx := r.Context()

	if err := app.store.Books.Create(ctx, book); err != nil {
		if errors.Is(err, store.ErrAuthorNotFound) {
			app.badRequestError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	w.Header().Set("Location", "/v1/books/"+strconv.FormatInt(book.ID, 10))

	if err := writeJSON(w, http.StatusCreated, book); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

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

	if err := writeJSON(w, http.StatusOK, book); err != nil {
		app.internalServerError(w, r, err)
	}
}

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

	if payload.Title != nil {
		book.Title = *payload.Title
	}
	if payload.Description != nil {
		book.Description = *payload.Description
	}
	if payload.AuthorID != nil {
		book.AuthorID = *payload.AuthorID
	}

	if err := app.store.Books.Update(r.Context(), book); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		case errors.Is(err, store.ErrAuthorNotFound):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, book); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := app.store.Books.GetAll(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, books); err != nil {
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
