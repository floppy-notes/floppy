package item

import "testing"

func TestIsValidItemType(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"note", "note", true},
		{"task", "task", true},
		{"reminder", "reminder", true},
		{"unknown", "event", false},
		{"empty", "", false},
		{"wrong case", "Note", false},
		{"with spaces", " note ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidItemType(tt.in)

			if got != tt.want {
				t.Errorf("IsValidItemType(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsValidTaskStatus(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"open", "open", true},
		{"done", "done", true},
		{"archived", "archived", true},
		{"reminder status", "pending", false},
		{"unknown", "sleeping", false},
		{"empty", "", false},
		{"wrong case", "Open", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTaskStatus(tt.in)

			if got != tt.want {
				t.Errorf("IsValidTaskStatus(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsValidReminderStatus(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"pending", "pending", true},
		{"fired", "fired", true},
		{"dismissed", "dismissed", true},
		{"task status", "open", false},
		{"unknown", "sleeping", false},
		{"empty", "", false},
		{"wrong case", "Pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidReminderStatus(tt.in)

			if got != tt.want {
				t.Errorf("IsValidReminderStatus(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
