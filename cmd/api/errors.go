package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = writeAPIError(w, http.StatusInternalServerError, "internal_error", "the server encountered a problem; try again later", nil)
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = writeAPIError(w, http.StatusNotFound, "not_found", "resource not found", nil)
}

func (app *application) validationError(w http.ResponseWriter, r *http.Request, fields map[string]string) {
	log.Printf("validation error: %s path: %s fields: %v", r.Method, r.URL.Path, fields)
	_ = writeAPIError(w, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", fields)
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = writeAPIError(w, http.StatusConflict, "conflict", err.Error(), nil)
}
