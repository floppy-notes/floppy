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
