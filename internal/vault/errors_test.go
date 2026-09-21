package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func readOnlyDir(t *testing.T) string {
	t.Helper()

	if os.Getuid() == 0 {
		t.Skip("running as root, permission bits are not enforced")
	}

	dir := filepath.Join(t.TempDir(), "locked")

	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatalf("Mkdir() returned unexpected error: %v", err)
	}

	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	return dir
}

func TestWriteErrors(t *testing.T) {
	t.Run("should fail when the parent folder cannot be created", func(t *testing.T) {
		path := filepath.Join(readOnlyDir(t), "nested", "note.md")

		err := Write(path, []byte("content"))

		if err == nil {
			t.Fatal("Write() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "creating folder for") {
			t.Errorf("Write() error = %q, want prefix %q", err.Error(), "creating folder for")
		}
	})

	t.Run("should fail when the file cannot be opened", func(t *testing.T) {
		path := filepath.Join(readOnlyDir(t), "note.md")

		err := Write(path, []byte("content"))

		if err == nil {
			t.Fatal("Write() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "writing") {
			t.Errorf("Write() error = %q, want prefix %q", err.Error(), "writing")
		}
	})
}

func TestOverwriteErrors(t *testing.T) {
	t.Run("should fail when the parent folder cannot be created", func(t *testing.T) {
		path := filepath.Join(readOnlyDir(t), "nested", "note.md")

		err := Overwrite(path, []byte("content"))

		if err == nil {
			t.Fatal("Overwrite() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "creating folder for") {
			t.Errorf("Overwrite() error = %q, want prefix %q", err.Error(), "creating folder for")
		}
	})

	t.Run("should fail when the file cannot be opened", func(t *testing.T) {
		path := filepath.Join(readOnlyDir(t), "note.md")

		err := Overwrite(path, []byte("content"))

		if err == nil {
			t.Fatal("Overwrite() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "writing") {
			t.Errorf("Overwrite() error = %q, want prefix %q", err.Error(), "writing")
		}
	})
}

func TestBuildIdErrors(t *testing.T) {
	t.Run("should fail when the sequence cannot be read", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "notes")

		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := BuildId(blocker, "tech sync", baseTime)

		if err == nil {
			t.Fatal("BuildId() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "building id") {
			t.Errorf("BuildId() error = %q, want prefix %q", err.Error(), "building id")
		}
	})
}

func TestWalkErrors(t *testing.T) {
	t.Run("should skip a file it cannot read", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, permission bits are not enforced")
		}

		root := t.TempDir()
		path := filepath.Join(root, "unreadable.md")

		if err := os.WriteFile(path, []byte(validNote), 0o000); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}
		t.Cleanup(func() { os.Chmod(path, 0o644) })

		items, skipped, err := Walk(root)

		if err != nil {
			t.Fatalf("Walk() returned unexpected error: %v", err)
		}

		if len(items) != 0 {
			t.Errorf("Walk() items = %d, want %d", len(items), 0)
		}

		if len(skipped) != 1 {
			t.Fatalf("Walk() skipped = %d, want %d", len(skipped), 1)
		}

		if !strings.Contains(skipped[0].Error(), "unreadable.md") {
			t.Errorf("skipped error = %q, want it to contain %q", skipped[0].Error(), "unreadable.md")
		}
	})
}

func TestNextIdSequenceErrors(t *testing.T) {
	t.Run("should fail when the path is not a folder", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "a-file")

		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := nextIdSequence(filepath.Join(blocker, "2026", "08"))

		if err == nil {
			t.Fatal("nextIdSequence() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "walking") {
			t.Errorf("nextIdSequence() error = %q, want prefix %q", err.Error(), "walking")
		}
	})

	t.Run("should still build an id after a valid month folder", func(t *testing.T) {
		folder := t.TempDir()

		got, err := BuildId(folder, "tech sync", time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC))

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		if got != "2027-01-05-0001-tech-sync" {
			t.Errorf("BuildId() = %q, want %q", got, "2027-01-05-0001-tech-sync")
		}
	})
}
