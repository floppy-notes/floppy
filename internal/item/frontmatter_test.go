package item

import (
	"testing"
)

const validCreated = "2026-08-12T00:00:00-03:00"

func TestFrontmatterValidate(t *testing.T) {
	tests := []struct {
		name    string
		front   Frontmatter
		wantErr string
	}{
		// valid
		{
			"valid note",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note", Created: validCreated},
			"",
		},
		{
			"valid task",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Due: "2026-08-20", Status: "open"},
			"",
		},
		{
			"valid reminder",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, RemindAt: validCreated, Status: "pending"},
			"",
		},

		// general
		{
			"empty id",
			Frontmatter{Title: "tech sync", Type: "note", Created: validCreated},
			"id can't be empty",
		},
		{
			"whitespace id",
			Frontmatter{ID: "   ", Title: "tech sync", Type: "note", Created: validCreated},
			"id can't be empty",
		},
		{
			"empty title",
			Frontmatter{ID: "id", Type: "note", Created: validCreated},
			"title can't be empty",
		},
		{
			"whitespace title",
			Frontmatter{ID: "id", Title: "   ", Type: "note", Created: validCreated},
			"title can't be empty",
		},
		{
			"invalid type",
			Frontmatter{ID: "id", Title: "tech sync", Type: "event", Created: validCreated},
			"invalid type: event (should be: note, task or reminder)",
		},
		{
			"empty type",
			Frontmatter{ID: "id", Title: "tech sync", Created: validCreated},
			"invalid type:  (should be: note, task or reminder)",
		},
		{
			"invalid created format",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note", Created: "29/05/2026"},
			"invalid created date format: 29/05/2026 (should be: 2006-01-02T15:04:05Z07:00)",
		},
		{
			"empty created",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note"},
			"invalid created date format:  (should be: 2006-01-02T15:04:05Z07:00)",
		},

		// note
		{
			"note with due",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note", Created: validCreated, Due: "2026-08-20"},
			"note must not have a due date",
		},
		{
			"note with status",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note", Created: validCreated, Status: "open"},
			"note must not have a status",
		},
		{
			"note with remind_at",
			Frontmatter{ID: "id", Title: "tech sync", Type: "note", Created: validCreated, RemindAt: validCreated},
			"note must not have a remind_at",
		},

		// task
		{
			"task with invalid due format",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Due: "2026-08-20T00:00:00", Status: "open"},
			"invalid due date format: 2026-08-20T00:00:00 (should be: 2006-01-02)",
		},
		{
			"task without due",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Status: "open"},
			"invalid due date format:  (should be: 2006-01-02)",
		},
		{
			"task with reminder status",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Due: "2026-08-20", Status: "pending"},
			"invalid task status: pending (should be: open, done, archived)",
		},
		{
			"task without status",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Due: "2026-08-20"},
			"invalid task status:  (should be: open, done, archived)",
		},
		{
			"task with remind_at",
			Frontmatter{ID: "id", Title: "write tests", Type: "task", Created: validCreated, Due: "2026-08-20", Status: "open", RemindAt: validCreated},
			"task must not have a remind_at",
		},

		// reminder
		{
			"reminder with invalid remind_at format",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, RemindAt: "2026-08-20", Status: "pending"},
			`invalid remind_at date format: "2026-08-20" (should be: 2006-01-02T15:04:05Z07:00)`,
		},
		{
			"reminder without remind_at",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, Status: "pending"},
			`invalid remind_at date format: "" (should be: 2006-01-02T15:04:05Z07:00)`,
		},
		{
			"reminder with task status",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, RemindAt: validCreated, Status: "open"},
			`invalid reminder status: "open" (should be: pending, fired, dismissed)`,
		},
		{
			"reminder without status",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, RemindAt: validCreated},
			`invalid reminder status: "" (should be: pending, fired, dismissed)`,
		},
		{
			"reminder with due",
			Frontmatter{ID: "id", Title: "call back", Type: "reminder", Created: validCreated, RemindAt: validCreated, Status: "pending", Due: "2026-08-20"},
			"reminder must not have a due date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.front.Validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() returned unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("Validate() returned nil error, want an error")
			}

			if err.Error() != tt.wantErr {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
