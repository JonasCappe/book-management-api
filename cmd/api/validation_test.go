package main

import "testing"

func TestNotBlankValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "non-blank value",
			value:   "The Hobbit",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate.Var(tt.value, "notblank")

			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate.Var() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
