package main

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(
		"internal server error",
		"request_id", middleware.GetReqID(r.Context()),
		"client_ip", middleware.GetClientIP(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)
	_ = writeAPIError(w, http.StatusInternalServerError, "internal_error", "the server encountered a problem; try again later", nil)
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn(
		"bad request",
		"request_id", middleware.GetReqID(r.Context()),
		"client_ip", middleware.GetClientIP(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)
	_ = writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Info(
		"resource not found",
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)
	_ = writeAPIError(w, http.StatusNotFound, "not_found", "resource not found", nil)
}

func (app *application) validationError(w http.ResponseWriter, r *http.Request, fields map[string]string) {
	app.logger.Warn(
		"validation failed",
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"fields", fields,
	)
	_ = writeAPIError(w, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", fields)
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn(
		"internal server error",
		"request_id", middleware.GetReqID(r.Context()),
		"client_ip", middleware.GetClientIP(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)
	_ = writeAPIError(w, http.StatusConflict, "conflict", err.Error(), nil)
}
