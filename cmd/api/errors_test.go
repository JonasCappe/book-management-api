package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestWriteAPIError(t *testing.T) {
	recorder := httptest.NewRecorder()
	fields := map[string]string{"title": "is required"}
	if err := writeAPIError(recorder, 422, "validation_failed", "request validation failed", fields); err != nil {
		t.Fatalf("writeAPIError() error = %v", err)
	}

	var response struct {
		Error struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Fields  map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != "validation_failed" || response.Error.Message != "request validation failed" {
		t.Fatalf("unexpected error response: %#v", response.Error)
	}
	if response.Error.Fields["title"] != "is required" {
		t.Fatalf("title error = %q", response.Error.Fields["title"])
	}
}
