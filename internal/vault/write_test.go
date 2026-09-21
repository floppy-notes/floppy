package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Run("should create the file and its parent folders", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "notes", "2026", "08", "note.md")

		if err := Write(path, []byte("content")); err != nil {
			t.Fatalf("Write() returned unexpected error: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if string(got) != "content" {
			t.Errorf("file content = %q, want %q", string(got), "content")
		}
	})

	t.Run("should fail when the file already exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")

		if err := Write(path, []byte("first")); err != nil {
			t.Fatalf("Write() returned unexpected error: %v", err)
		}

		err := Write(path, []byte("second"))
		if err == nil {
			t.Fatal("Write() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "writing") {
			t.Errorf("Write() error = %q, want prefix %q", err.Error(), "writing")
		}
	})

	t.Run("should keep the original content when the file already exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")

		if err := Write(path, []byte("first")); err != nil {
			t.Fatalf("Write() returned unexpected error: %v", err)
		}

		if err := Write(path, []byte("second")); err == nil {
			t.Fatal("Write() returned nil error, want an error")
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if string(got) != "first" {
			t.Errorf("file content = %q, want %q", string(got), "first")
		}
	})
}

func TestOverwrite(t *testing.T) {
	t.Run("should create the file and its parent folders", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "notes", "2026", "08", "note.md")

		if err := Overwrite(path, []byte("content")); err != nil {
			t.Fatalf("Overwrite() returned unexpected error: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if string(got) != "content" {
			t.Errorf("file content = %q, want %q", string(got), "content")
		}
	})

	t.Run("should replace the content of an existing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")

		if err := Overwrite(path, []byte("first")); err != nil {
			t.Fatalf("Overwrite() returned unexpected error: %v", err)
		}

		if err := Overwrite(path, []byte("second")); err != nil {
			t.Fatalf("Overwrite() returned unexpected error: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if string(got) != "second" {
			t.Errorf("file content = %q, want %q", string(got), "second")
		}
	})

	t.Run("should truncate when the new content is shorter", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")

		if err := Overwrite(path, []byte("a very long content")); err != nil {
			t.Fatalf("Overwrite() returned unexpected error: %v", err)
		}

		if err := Overwrite(path, []byte("short")); err != nil {
			t.Fatalf("Overwrite() returned unexpected error: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() returned unexpected error: %v", err)
		}

		if string(got) != "short" {
			t.Errorf("file content = %q, want %q", string(got), "short")
		}
	})
}
