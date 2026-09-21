package item

import (
	"reflect"
	"strings"
	"testing"
)

var baseItem = Item{
	Front: Frontmatter{
		ID:      "2026-08-12-0011-example",
		Title:   "tech sync",
		Type:    "note",
		Created: "2026-08-12T00:00:00-03:00",
		Tags:    []string{"tech", "meetings"},
		Related: []string{"tech-group"},
		Extra:   map[string]any{"keep": "me"},
	},
	Body: "Lorem ipsum dolor sit amet. Cum ipsum repellendus aut ipsam voluptatem hic rerum tempore qui galisum quos sed nihil exercitationem id galisum commodi eum ducimus enim. In optio repudiandae vel repellendus dolore qui internos odit.",
}

var taskItem = Item{
	Front: Frontmatter{
		ID:      "2026-08-12-0012-task",
		Title:   "write tests",
		Type:    "task",
		Created: "2026-08-12T00:00:00-03:00",
		Due:     "2026-08-20",
		Status:  "open",
	},
	Body: "write the format round-trip tests",
}

var reminderItem = Item{
	Front: Frontmatter{
		ID:       "2026-08-12-0013-reminder",
		Title:    "call back",
		Type:     "reminder",
		Created:  "2026-08-12T00:00:00-03:00",
		RemindAt: "2026-08-20T10:00:00-03:00",
		Status:   "pending",
	},
	Body: "call the client back",
}

var minimalItem = Item{
	Front: Frontmatter{
		ID:      "2026-08-12-0017-minimal",
		Title:   "bare note",
		Type:    "note",
		Created: "2026-08-12T00:00:00-03:00",
	},
	Body: "just a body",
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		in   Item
	}{
		{"note with a custom frontmatter field", baseItem},
		{"task with due date and status", taskItem},
		{"reminder with remind_at and status", reminderItem},
		{"note with no optional fields", minimalItem},
	}

	for _, tt := range tests {
		t.Run("should round-trip a "+tt.name, func(t *testing.T) {
			raw, err := Format(tt.in)
			if err != nil {
				t.Fatalf("Format() returned unexpected error: %v", err)
			}

			parsedItem, err := Parse(raw)
			if err != nil {
				t.Fatalf("Parse() returned unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.in, parsedItem) {
				t.Errorf("round-trip mismatch: got %#v, want %#v", parsedItem, tt.in)
			}
		})
	}
}

func TestFormatInvalid(t *testing.T) {
	tests := []struct {
		name       string
		in         Item
		wantReason string
	}{
		{
			"note with a due date",
			Item{Front: Frontmatter{
				ID:      "2026-08-12-0014-invalid",
				Title:   "note with a due date",
				Type:    "note",
				Created: "2026-08-12T00:00:00-03:00",
				Due:     "2026-08-20",
			}},
			"note must not have a due date",
		},
		{
			"task with an unknown status",
			Item{Front: Frontmatter{
				ID:      "2026-08-12-0015-invalid-task",
				Title:   "broken task",
				Type:    "task",
				Created: "2026-08-12T00:00:00-03:00",
				Due:     "2026-08-20",
				Status:  "not-a-status",
			}},
			"invalid task status",
		},
		{
			"reminder with a due date",
			Item{Front: Frontmatter{
				ID:       "2026-08-12-0016-invalid-reminder",
				Title:    "broken reminder",
				Type:     "reminder",
				Created:  "2026-08-12T00:00:00-03:00",
				RemindAt: "2026-08-20T10:00:00-03:00",
				Status:   "pending",
				Due:      "2026-08-20",
			}},
			"reminder must not have a due date",
		},
	}

	wantPrefix := "invalid item"

	for _, tt := range tests {
		t.Run("should fail on a "+tt.name, func(t *testing.T) {
			_, err := Format(tt.in)

			if err == nil {
				t.Fatal("Format() returned nil error, want an error")
			}

			if !strings.HasPrefix(err.Error(), wantPrefix) {
				t.Errorf("Format() error = %q, want prefix %q", err.Error(), wantPrefix)
			}

			if !strings.Contains(err.Error(), tt.wantReason) {
				t.Errorf("Format() error = %q, want it to contain %q", err.Error(), tt.wantReason)
			}
		})
	}
}
