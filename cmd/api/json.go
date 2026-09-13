package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // 1 MiB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		var syntaxError *json.SyntaxError
		var typeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains malformed JSON at position %d", syntaxError.Offset)
		case errors.As(err, &typeError):
			return fmt.Errorf("request body contains an invalid value for %q", typeError.Field)
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not exceed %d bytes", maxBytes)
		default:
			return err
		}
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, msg string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(w, status, &envelope{Error: msg})
}
