package data

import (
	"encoding/json"
	"testing"
)

func TestDateJSONRoundTrip(t *testing.T) {
	original, err := ParseDate("2020-05-17")
	if err != nil {
		t.Fatalf("ParseDate() error = %v", err)
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(encoded) != `"2020-05-17"` {
		t.Fatalf("json.Marshal() = %s, want %q", encoded, `"2020-05-17"`)
	}

	var decoded Date
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !decoded.Equal(original.Time) {
		t.Fatalf("decoded date = %s, want %s", decoded.Time, original.Time)
	}
}

func TestDateRejectsInvalidFormat(t *testing.T) {
	var date Date
	if err := json.Unmarshal([]byte(`"17-05-2020"`), &date); err == nil {
		t.Fatal("json.Unmarshal() error = nil, want invalid date error")
	}
}
