package cli

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"date only", "2026-08-12", time.Date(2026, 8, 12, 0, 0, 0, 0, time.Local)},
		{"date and time with T", "2026-08-12T10:30", time.Date(2026, 8, 12, 10, 30, 0, 0, time.Local)},
		{"date and time with space", "2026-08-12 10:30", time.Date(2026, 8, 12, 10, 30, 0, 0, time.Local)},
		{"empty returns the zero time", "", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.in)

			if err != nil {
				t.Fatalf("parseDate() returned unexpected error: %v", err)
			}

			if !got.Equal(tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDateInvalid(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"brazilian format", "12/08/2026"},
		{"full rfc3339", "2026-08-12T10:30:00-03:00"},
		{"seconds are not supported", "2026-08-12 10:30:00"},
		{"impossible month", "2026-13-01"},
		{"not a date", "tomorrow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.in)

			if err == nil {
				t.Fatal("parseDate() returned nil error, want an error")
			}

			if !got.IsZero() {
				t.Errorf("parseDate() = %v, want the zero time", got)
			}
		})
	}
}
