package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JonasCappe/book-management-api/internal/data"
	"github.com/JonasCappe/book-management-api/internal/store"
)

// getBookHistoryHandler godoc
//
//	@Summary		List book history
//	@Description	Returns paginated change history for a book, including soft-deleted books.
//	@Tags			history
//	@Produce		json
//	@Param			bookID		path		int		true	"Book ID"
//	@Param			page		query		int		false	"Page number"			default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"		default(20)	minimum(1)	maximum(100)
//	@Param			change_type	query		string	false	"Change type"			Enums(created, updated, deleted)
//	@Param			order		query		string	false	"Chronological order"	default(desc)	Enums(asc, desc)
//	@Success		200			{object}	HistoryResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/v1/books/{bookID}/history [get]
func (app *application) getBookHistoryHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := app.readBookID(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	filters, err := readHistoryQuery(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	page, err := app.store.HistoryEntries.GetByBookID(r.Context(), bookID, filters)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, page); err != nil {
		app.internalServerError(w, r, err)
	}
}

func readHistoryQuery(r *http.Request) (data.HistoryQuery, error) {
	query := data.HistoryQuery{Page: 1, PageSize: 20, Order: "desc"}
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
	if value := values.Get("change_type"); value != "" {
		query.ChangeType = data.ChangeType(value)
		switch query.ChangeType {
		case data.ChangeTypeCreated, data.ChangeTypeUpdated, data.ChangeTypeDeleted:
		default:
			return query, fmt.Errorf("change_type must be created, updated, or deleted")
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
