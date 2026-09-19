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

func TestFormat(t *testing.T) {
	t.Run("should round-trip a note with a custom frontmatter field", func(t *testing.T) {
		raw, err := Format(baseItem)
		if err != nil {
			t.Fatalf("Format() returned unexpected error: %v", err)
		}

		parsedItem, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(baseItem, parsedItem) {
			t.Errorf("round-trip mismatch: got %#v, want %#v", parsedItem, baseItem)
		}
	})

	t.Run("should round-trip a task with due date and status", func(t *testing.T) {
		raw, err := Format(taskItem)
		if err != nil {
			t.Fatalf("Format() returned unexpected error: %v", err)
		}

		parsedItem, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(taskItem, parsedItem) {
			t.Errorf("round-trip mismatch: got %#v, want %#v", parsedItem, taskItem)
		}
	})

	t.Run("should round-trip a reminder with remind_at and status", func(t *testing.T) {
		raw, err := Format(reminderItem)
		if err != nil {
			t.Fatalf("Format() returned unexpected error: %v", err)
		}

		parsedItem, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(reminderItem, parsedItem) {
			t.Errorf("round-trip mismatch: got %#v, want %#v", parsedItem, reminderItem)
		}
	})

	t.Run("should round-trip a note with no optional fields", func(t *testing.T) {
		minimalItem := Item{
			Front: Frontmatter{
				ID:      "2026-08-12-0017-minimal",
				Title:   "bare note",
				Type:    "note",
				Created: "2026-08-12T00:00:00-03:00",
			},
			Body: "just a body",
		}

		raw, err := Format(minimalItem)
		if err != nil {
			t.Fatalf("Format() returned unexpected error: %v", err)
		}

		parsedItem, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(minimalItem, parsedItem) {
			t.Errorf("round-trip mismatch: got %#v, want %#v", parsedItem, minimalItem)
		}
	})

	t.Run("should fail when the note is invalid", func(t *testing.T) {
		invalidItem := Item{
			Front: Frontmatter{
				ID:      "2026-08-12-0014-invalid",
				Title:   "note with a due date",
				Type:    "note",
				Created: "2026-08-12T00:00:00-03:00",
				Due:     "2026-08-20",
			},
		}

		wantPrefix := "invalid item"
		wantReason := "note must not have a due date"

		_, err := Format(invalidItem)
		if err == nil {
			t.Fatal("Format() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), wantPrefix) {
			t.Errorf("Format() error = %q, want prefix %q", err.Error(), wantPrefix)
		}

		if !strings.Contains(err.Error(), wantReason) {
			t.Errorf("Format() error = %q, want it to contain %q", err.Error(), wantReason)
		}
	})

	t.Run("should fail when the task is invalid", func(t *testing.T) {
		invalidItem := Item{
			Front: Frontmatter{
				ID:      "2026-08-12-0015-invalid-task",
				Title:   "broken task",
				Type:    "task",
				Created: "2026-08-12T00:00:00-03:00",
				Due:     "2026-08-20",
				Status:  "not-a-status",
			},
		}

		wantPrefix := "invalid item"
		wantReason := "invalid task status"

		_, err := Format(invalidItem)
		if err == nil {
			t.Fatal("Format() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), wantPrefix) {
			t.Errorf("Format() error = %q, want prefix %q", err.Error(), wantPrefix)
		}

		if !strings.Contains(err.Error(), wantReason) {
			t.Errorf("Format() error = %q, want it to contain %q", err.Error(), wantReason)
		}
	})

	t.Run("should fail when the reminder is invalid", func(t *testing.T) {
		invalidItem := Item{
			Front: Frontmatter{
				ID:       "2026-08-12-0016-invalid-reminder",
				Title:    "broken reminder",
				Type:     "reminder",
				Created:  "2026-08-12T00:00:00-03:00",
				RemindAt: "2026-08-20T10:00:00-03:00",
				Status:   "pending",
				Due:      "2026-08-20",
			},
		}

		wantPrefix := "invalid item"
		wantReason := "reminder must not have a due date"

		_, err := Format(invalidItem)
		if err == nil {
			t.Fatal("Format() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), wantPrefix) {
			t.Errorf("Format() error = %q, want prefix %q", err.Error(), wantPrefix)
		}

		if !strings.Contains(err.Error(), wantReason) {
			t.Errorf("Format() error = %q, want it to contain %q", err.Error(), wantReason)
		}
	})
}
