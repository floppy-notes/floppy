package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var baseTime = time.Date(2026, 8, 12, 9, 30, 0, 0, time.UTC)

func seedItem(t *testing.T, folder string, id string) {
	t.Helper()

	path := filepath.Join(folder, id+".md")

	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned unexpected error: %v", err)
	}

	content := []byte("---\nid: " + id + "\ntitle: tech sync\ntype: note\ncreated: \"2026-08-12T00:00:00-03:00\"\n---\n\nbody\n")

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() returned unexpected error: %v", err)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase and join words", "Tech Sync", "tech-sync"},
		{"strip accents", "Reunião Semanal", "reuniao-semanal"},
		{"strip cedilla", "Manutenção", "manutencao"},
		{"collapse repeated symbols", "a -- b", "a-b"},
		{"trim leading and trailing dashes", "--hello--", "hello"},
		{"keep digits", "Sprint 42", "sprint-42"},
		{"drop emoji", "deploy 🚀 hoje", "deploy-hoje"},
		{"untitled when nothing remains", "!!! ???", "untitled"},
		{"untitled when empty", "", "untitled"},
		{"untitled when only spaces", "   ", "untitled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.in)

			if got != tt.want {
				t.Errorf("slugify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextIdSequence(t *testing.T) {
	t.Run("should return 1 when the folder does not exist", func(t *testing.T) {
		missingFolder := filepath.Join(t.TempDir(), "notes", "2026", "08")

		got, err := nextIdSequence(missingFolder)

		if err != nil {
			t.Fatalf("nextIdSequence() returned unexpected error: %v", err)
		}

		if got != 1 {
			t.Errorf("nextIdSequence() = %d, want %d", got, 1)
		}
	})

	t.Run("should return 1 when the folder is empty", func(t *testing.T) {
		emptyFolder := t.TempDir()

		got, err := nextIdSequence(emptyFolder)

		if err != nil {
			t.Fatalf("nextIdSequence() returned unexpected error: %v", err)
		}

		if got != 1 {
			t.Errorf("nextIdSequence() = %d, want %d", got, 1)
		}
	})

	t.Run("should return the highest sequence plus one", func(t *testing.T) {
		folder := t.TempDir()
		seedItem(t, folder, "2026-08-12-0001-first")
		seedItem(t, folder, "2026-08-12-0007-second")
		seedItem(t, folder, "2026-08-12-0003-third")

		got, err := nextIdSequence(folder)

		if err != nil {
			t.Fatalf("nextIdSequence() returned unexpected error: %v", err)
		}

		if got != 8 {
			t.Errorf("nextIdSequence() = %d, want %d", got, 8)
		}
	})

	t.Run("should ignore ids with too few parts", func(t *testing.T) {
		folder := t.TempDir()
		seedItem(t, folder, "short-id")
		seedItem(t, folder, "2026-08-12-0002-valid")

		got, err := nextIdSequence(folder)

		if err != nil {
			t.Fatalf("nextIdSequence() returned unexpected error: %v", err)
		}

		if got != 3 {
			t.Errorf("nextIdSequence() = %d, want %d", got, 3)
		}
	})

	t.Run("should ignore ids with a non numeric sequence", func(t *testing.T) {
		folder := t.TempDir()
		seedItem(t, folder, "2026-08-12-abcd-broken")
		seedItem(t, folder, "2026-08-12-0005-valid")

		got, err := nextIdSequence(folder)

		if err != nil {
			t.Fatalf("nextIdSequence() returned unexpected error: %v", err)
		}

		if got != 6 {
			t.Errorf("nextIdSequence() = %d, want %d", got, 6)
		}
	})
}

func TestBuildId(t *testing.T) {
	t.Run("should build the first id of the month", func(t *testing.T) {
		folder := t.TempDir()

		got, err := BuildId(folder, "Tech Sync", baseTime)

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0001-tech-sync"
		if got != want {
			t.Errorf("BuildId() = %q, want %q", got, want)
		}
	})

	t.Run("should continue the sequence of the same month", func(t *testing.T) {
		folder := t.TempDir()
		seedItem(t, filepath.Join(folder, "2026", "08"), "2026-08-12-0004-existing")

		got, err := BuildId(folder, "Tech Sync", baseTime)

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0005-tech-sync"
		if got != want {
			t.Errorf("BuildId() = %q, want %q", got, want)
		}
	})

	t.Run("should not be affected by items of another month", func(t *testing.T) {
		folder := t.TempDir()
		seedItem(t, filepath.Join(folder, "2026", "07"), "2026-07-12-0009-other-month")

		got, err := BuildId(folder, "Tech Sync", baseTime)

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0001-tech-sync"
		if got != want {
			t.Errorf("BuildId() = %q, want %q", got, want)
		}
	})

	t.Run("should slugify the title", func(t *testing.T) {
		folder := t.TempDir()

		got, err := BuildId(folder, "Reunião Semanal", baseTime)

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0001-reuniao-semanal"
		if got != want {
			t.Errorf("BuildId() = %q, want %q", got, want)
		}
	})

	t.Run("should fall back to untitled when the title has no usable characters", func(t *testing.T) {
		folder := t.TempDir()

		got, err := BuildId(folder, "!!!", baseTime)

		if err != nil {
			t.Fatalf("BuildId() returned unexpected error: %v", err)
		}

		want := "2026-08-12-0001-untitled"
		if got != want {
			t.Errorf("BuildId() = %q, want %q", got, want)
		}
	})
}
