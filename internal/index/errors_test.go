package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/item"
)

func TestCreateErrors(t *testing.T) {
	t.Run("should fail when the id cannot be built", func(t *testing.T) {
		root := t.TempDir()
		blocker := filepath.Join(root, "notes")

		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := Create(root, baseRequest, baseTime)

		if err == nil {
			t.Fatal("Create() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "building id") {
			t.Errorf("Create() error = %q, want prefix %q", err.Error(), "building id")
		}
	})

	t.Run("should fail when the file already exists", func(t *testing.T) {
		root := t.TempDir()

		created, err := Create(root, baseRequest, baseTime)
		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		if err := os.WriteFile(created.Path, []byte("squatter"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		req := baseRequest
		req.Title = "another title"

		second, err := Create(root, req, baseTime)
		if err != nil {
			t.Fatalf("Create() returned unexpected error: %v", err)
		}

		if second.Front.ID == created.Front.ID {
			t.Errorf("Create() reused the id %q, want a new one", second.Front.ID)
		}
	})
}

func TestUpdateErrors(t *testing.T) {
	t.Run("should fail when the file cannot be overwritten", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, permission bits are not enforced")
		}

		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := os.Chmod(seeded.Path, 0o444); err != nil {
			t.Fatalf("Chmod() returned unexpected error: %v", err)
		}
		t.Cleanup(func() { os.Chmod(seeded.Path, 0o644) })

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "writing") {
			t.Errorf("Update() error = %q, want prefix %q", err.Error(), "writing")
		}
	})

	t.Run("should fail when the index cannot be written", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if err := db.Conn.Close(); err != nil {
			t.Fatalf("Close() returned unexpected error: %v", err)
		}

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}
	})
}

func TestUpdateReminder(t *testing.T) {
	t.Run("should update remind_at and status", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeReminder
		req.RemindAt = "2026-08-20T10:00:00-03:00"
		seeded := seedItem(t, db, dir, req)

		got, err := Update(db, UpdateRequest{
			ID:       seeded.Front.ID,
			Type:     item.TypeReminder,
			RemindAt: "2026-09-01T08:15:00-03:00",
			Status:   "fired",
		}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		if got.Front.RemindAt != "2026-09-01T08:15:00-03:00" {
			t.Errorf("remind_at = %q, want %q", got.Front.RemindAt, "2026-09-01T08:15:00-03:00")
		}

		if got.Front.Status != "fired" {
			t.Errorf("status = %q, want %q", got.Front.Status, "fired")
		}
	})

	t.Run("should keep remind_at when it is not informed", func(t *testing.T) {
		db, dir := newTestDB(t)

		req := baseRequest
		req.Type = item.TypeReminder
		req.RemindAt = "2026-08-20T10:00:00-03:00"
		seeded := seedItem(t, db, dir, req)

		got, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeReminder, Status: "dismissed"}, false)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		if got.Front.RemindAt != "2026-08-20T10:00:00-03:00" {
			t.Errorf("remind_at = %q, want %q", got.Front.RemindAt, "2026-08-20T10:00:00-03:00")
		}
	})

	t.Run("should replace related when replacing", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		got, err := Update(db, UpdateRequest{
			ID:      seeded.Front.ID,
			Type:    item.TypeNote,
			Related: []string{"only-this"},
		}, true)

		if err != nil {
			t.Fatalf("Update() returned unexpected error: %v", err)
		}

		if len(got.Front.Related) != 1 || got.Front.Related[0] != "only-this" {
			t.Errorf("related = %#v, want %#v", got.Front.Related, []string{"only-this"})
		}
	})

	t.Run("should fail when the reindex cannot run", func(t *testing.T) {
		db, dir := newTestDB(t)
		seeded := seedItem(t, db, dir, baseRequest)

		if _, err := db.Conn.Exec("DROP TABLE tags"); err != nil {
			t.Fatalf("dropping tags returned unexpected error: %v", err)
		}

		_, err := Update(db, UpdateRequest{ID: seeded.Front.ID, Type: item.TypeNote, Title: "new title"}, false)

		if err == nil {
			t.Fatal("Update() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "deleting item") {
			t.Errorf("Update() error = %q, want it to contain %q", err.Error(), "deleting item")
		}
	})
}

func TestCreateWriteError(t *testing.T) {
	t.Run("should fail when the file cannot be written", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, permission bits are not enforced")
		}

		root := t.TempDir()
		monthDir := filepath.Join(root, "notes", "2026", "08")

		if err := os.MkdirAll(monthDir, 0o755); err != nil {
			t.Fatalf("MkdirAll() returned unexpected error: %v", err)
		}

		if err := os.Chmod(monthDir, 0o555); err != nil {
			t.Fatalf("Chmod() returned unexpected error: %v", err)
		}
		t.Cleanup(func() { os.Chmod(monthDir, 0o755) })

		_, err := Create(root, baseRequest, baseTime)

		if err == nil {
			t.Fatal("Create() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "writing") {
			t.Errorf("Create() error = %q, want prefix %q", err.Error(), "writing")
		}
	})
}
