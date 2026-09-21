package cli

import (
	"strings"
	"testing"
)

func seedNote(t *testing.T, title string) string {
	t.Helper()

	out, err := runCmd(t, newCreateNoteCmd(), "--title", title, "--body", "original body", "--tags", "tech")
	if err != nil {
		t.Fatalf("seeding a note returned unexpected error: %v", err)
	}

	return decodeItem(t, out).ID
}

func seedTask(t *testing.T, title string) string {
	t.Helper()

	out, err := runCmd(t, newCreateTaskCmd(), "--title", title, "--due", "2026-08-20")
	if err != nil {
		t.Fatalf("seeding a task returned unexpected error: %v", err)
	}

	return decodeItem(t, out).ID
}

func seedReminder(t *testing.T, title string) string {
	t.Helper()

	out, err := runCmd(t, newCreateReminderCmd(), "--title", title, "--remind-at", "2026-08-20 10:00")
	if err != nil {
		t.Fatalf("seeding a reminder returned unexpected error: %v", err)
	}

	return decodeItem(t, out).ID
}

func TestNewUpdateNoteCmd(t *testing.T) {
	t.Run("should update the title", func(t *testing.T) {
		useTempVault(t)
		id := seedNote(t, "tech sync")

		out, err := runCmd(t, newUpdateNoteCmd(), "--id", id, "--title", "new title")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Title != "new title" {
			t.Errorf("title = %q, want %q", got.Title, "new title")
		}

		if got.ID != id {
			t.Errorf("id = %q, want %q", got.ID, id)
		}
	})

	t.Run("should print nothing when quiet", func(t *testing.T) {
		useTempVault(t)
		id := seedNote(t, "tech sync")

		out, err := runCmd(t, newUpdateNoteCmd(), "--id", id, "--title", "new title", "--quiet")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want %q", out, "")
		}
	})

	t.Run("should fail when the id does not exist", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newUpdateNoteCmd(), "--id", "missing", "--title", "new title")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "updating item") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "updating item")
		}
	})

	t.Run("should fail when id is missing", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newUpdateNoteCmd(), "--title", "new title")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "id") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "id")
		}
	})

	t.Run("should not duplicate a tag the note already has", func(t *testing.T) {
		useTempVault(t)
		id := seedNote(t, "tech sync")

		if _, err := runCmd(t, newUpdateNoteCmd(), "--id", id, "--tags", "tech"); err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if _, err := runCmd(t, newUpdateNoteCmd(), "--id", id, "--tags", "tech"); err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}
	})
}

func TestNewUpdateTaskCmd(t *testing.T) {
	t.Run("should update the status", func(t *testing.T) {
		useTempVault(t)
		id := seedTask(t, "write tests")

		out, err := runCmd(t, newUpdateTaskCmd(), "--id", id, "--status", "done")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Status == nil {
			t.Fatal("status = nil, want a value")
		}

		if *got.Status != "done" {
			t.Errorf("status = %q, want %q", *got.Status, "done")
		}
	})

	t.Run("should normalize the due date", func(t *testing.T) {
		useTempVault(t)
		id := seedTask(t, "write tests")

		out, err := runCmd(t, newUpdateTaskCmd(), "--id", id, "--due", "2026-09-01 15:30")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Due == nil {
			t.Fatal("due = nil, want a value")
		}

		if *got.Due != "2026-09-01" {
			t.Errorf("due = %q, want %q", *got.Due, "2026-09-01")
		}
	})

	t.Run("should keep the current due date when it is not informed", func(t *testing.T) {
		useTempVault(t)
		id := seedTask(t, "write tests")

		out, err := runCmd(t, newUpdateTaskCmd(), "--id", id, "--title", "new title")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Due == nil {
			t.Fatal("due = nil, want a value")
		}

		if *got.Due != "2026-08-20" {
			t.Errorf("due = %q, want %q", *got.Due, "2026-08-20")
		}
	})

	t.Run("should fail when the due date format is invalid", func(t *testing.T) {
		useTempVault(t)
		id := seedTask(t, "write tests")

		_, err := runCmd(t, newUpdateTaskCmd(), "--id", id, "--due", "20/08/2026")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "invalid --due") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "invalid --due")
		}
	})

	t.Run("should fail when the status belongs to a reminder", func(t *testing.T) {
		useTempVault(t)
		id := seedTask(t, "write tests")

		_, err := runCmd(t, newUpdateTaskCmd(), "--id", id, "--status", "pending")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "invalid task status") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "invalid task status")
		}
	})
}

func TestNewUpdateReminderCmd(t *testing.T) {
	t.Run("should update the status", func(t *testing.T) {
		useTempVault(t)
		id := seedReminder(t, "call back")

		out, err := runCmd(t, newUpdateReminderCmd(), "--id", id, "--status", "fired")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Status == nil {
			t.Fatal("status = nil, want a value")
		}

		if *got.Status != "fired" {
			t.Errorf("status = %q, want %q", *got.Status, "fired")
		}
	})

	t.Run("should normalize remind-at to RFC3339", func(t *testing.T) {
		useTempVault(t)
		id := seedReminder(t, "call back")

		out, err := runCmd(t, newUpdateReminderCmd(), "--id", id, "--remind-at", "2026-09-01 08:15")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.RemindAt == nil {
			t.Fatal("remind_at = nil, want a value")
		}

		if !strings.HasPrefix(*got.RemindAt, "2026-09-01T08:15:00") {
			t.Errorf("remind_at = %q, want prefix %q", *got.RemindAt, "2026-09-01T08:15:00")
		}
	})

	t.Run("should fail when the remind-at format is invalid", func(t *testing.T) {
		useTempVault(t)
		id := seedReminder(t, "call back")

		_, err := runCmd(t, newUpdateReminderCmd(), "--id", id, "--remind-at", "amanha")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "invalid --remind-at") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "invalid --remind-at")
		}
	})

	t.Run("should fail when the status belongs to a task", func(t *testing.T) {
		useTempVault(t)
		id := seedReminder(t, "call back")

		_, err := runCmd(t, newUpdateReminderCmd(), "--id", id, "--status", "done")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "invalid reminder status") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "invalid reminder status")
		}
	})

	t.Run("should replace the body when replacing", func(t *testing.T) {
		useTempVault(t)
		id := seedReminder(t, "call back")

		if _, err := runCmd(t, newUpdateReminderCmd(), "--id", id, "--body", "new body", "--replace"); err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}
	})
}
