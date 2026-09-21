package cli

import (
	"testing"
)

func TestListFlagsValidate(t *testing.T) {
	t.Run("should accept empty flags", func(t *testing.T) {
		flags := ListFlags{}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}
	})

	t.Run("should accept a valid type", func(t *testing.T) {
		flags := ListFlags{Type: "task"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}
	})

	t.Run("should return error when the type is invalid", func(t *testing.T) {
		flags := ListFlags{Type: "event"}

		wantMsg := "invalid item type: event"

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should accept a task status", func(t *testing.T) {
		flags := ListFlags{Status: "done"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}
	})

	t.Run("should accept a reminder status", func(t *testing.T) {
		flags := ListFlags{Status: "dismissed"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}
	})

	t.Run("should return error when the status does not exist", func(t *testing.T) {
		flags := ListFlags{Status: "sleeping"}

		wantMsg := "nonexistent status: sleeping"

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should normalize since to RFC3339", func(t *testing.T) {
		flags := ListFlags{Since: "2026-08-12"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}

		want := "2026-08-12T00:00:00"
		if len(flags.Since) < len(want) || flags.Since[:len(want)] != want {
			t.Errorf("validate() since = %q, want it to start with %q", flags.Since, want)
		}
	})

	t.Run("should normalize until to RFC3339", func(t *testing.T) {
		flags := ListFlags{Until: "2026-08-12 23:59"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}

		want := "2026-08-12T23:59:00"
		if len(flags.Until) < len(want) || flags.Until[:len(want)] != want {
			t.Errorf("validate() until = %q, want it to start with %q", flags.Until, want)
		}
	})

	t.Run("should return error when since is invalid", func(t *testing.T) {
		flags := ListFlags{Since: "12/08/2026"}

		wantMsg := `invalid since date: "12/08/2026": expected YYYY-MM-DD or YYYY-MM-DD HH:MM`

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should return error when until is invalid", func(t *testing.T) {
		flags := ListFlags{Until: "12/08/2026"}

		wantMsg := `invalid until date: "12/08/2026": expected YYYY-MM-DD or YYYY-MM-DD HH:MM`

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should reject a reminder status when filtering tasks", func(t *testing.T) {
		flags := ListFlags{Type: "task", Status: "pending"}

		wantMsg := "invalid task status: pending (should be: open, done, archived)"

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should reject a task status when filtering reminders", func(t *testing.T) {
		flags := ListFlags{Type: "reminder", Status: "open"}

		wantMsg := "invalid reminder status: open (should be: pending, fired, dismissed)"

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should reject any status when filtering notes", func(t *testing.T) {
		flags := ListFlags{Type: "note", Status: "open"}

		wantMsg := "notes do not have a status"

		err := flags.validate()
		if err == nil {
			t.Fatal("validate() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("validate() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should accept a status that matches the type", func(t *testing.T) {
		flags := ListFlags{Type: "task", Status: "done"}

		if err := flags.validate(); err != nil {
			t.Fatalf("validate() returned unexpected error: %v", err)
		}
	})
}
