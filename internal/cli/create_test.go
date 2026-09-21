package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func useTempVault(t *testing.T) string {
	t.Helper()

	original := vaultPath
	t.Cleanup(func() { vaultPath = original })

	vaultPath = t.TempDir()

	return vaultPath
}

func runCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer

	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()

	return out.String(), err
}

func decodeItem(t *testing.T, raw string) itemView {
	t.Helper()

	var view itemView
	if err := json.Unmarshal([]byte(raw), &view); err != nil {
		t.Fatalf("Unmarshal() returned unexpected error: %v, raw = %q", err, raw)
	}

	return view
}

func TestCreateFlagsValidate(t *testing.T) {
	tests := []struct {
		name    string
		flags   createFlags
		wantErr string
	}{
		{"title informed", createFlags{Title: "tech sync"}, ""},
		{"title with surrounding spaces", createFlags{Title: "  tech sync  "}, ""},
		{"empty title", createFlags{}, "title is required"},
		{"whitespace title", createFlags{Title: "   "}, "title is required"},
		{"tab title", createFlags{Title: "\t"}, "title is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.Validate()

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

func TestNewCreateNoteCmd(t *testing.T) {
	t.Run("should create a note and print it", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateNoteCmd(), "--title", "tech sync", "--body", "the body")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Title != "tech sync" {
			t.Errorf("title = %q, want %q", got.Title, "tech sync")
		}

		if got.Type != "note" {
			t.Errorf("type = %q, want %q", got.Type, "note")
		}

		if got.Status != nil {
			t.Errorf("status = %v, want nil", got.Status)
		}
	})

	t.Run("should print nothing when quiet", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateNoteCmd(), "--title", "tech sync", "--quiet")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want %q", out, "")
		}
	})

	t.Run("should fail when the title is missing", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newCreateNoteCmd(), "--body", "the body")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if err.Error() != "title is required" {
			t.Errorf("Execute() error = %q, want %q", err.Error(), "title is required")
		}
	})

	t.Run("should attach tags and related", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateNoteCmd(),
			"--title", "tech sync",
			"--tags", "tech,meetings",
			"--related", "other-id",
		)

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if !strings.Contains(out, "tech sync") {
			t.Errorf("output = %q, want it to contain %q", out, "tech sync")
		}
	})
}

func TestNewCreateTaskCmd(t *testing.T) {
	t.Run("should normalize the due date", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateTaskCmd(), "--title", "write tests", "--due", "2026-08-20")

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

	t.Run("should drop the time part of the due date", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateTaskCmd(), "--title", "write tests", "--due", "2026-08-20 15:30")

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

	t.Run("should default the status to open", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateTaskCmd(), "--title", "write tests", "--due", "2026-08-20")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Status == nil {
			t.Fatal("status = nil, want a value")
		}

		if *got.Status != "open" {
			t.Errorf("status = %q, want %q", *got.Status, "open")
		}
	})

	t.Run("should fail when the due date format is invalid", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newCreateTaskCmd(), "--title", "write tests", "--due", "20/08/2026")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "invalid --due") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "invalid --due")
		}
	})

	t.Run("should fail when due is missing", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newCreateTaskCmd(), "--title", "write tests")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "due") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "due")
		}
	})
}

func TestNewCreateReminderCmd(t *testing.T) {
	t.Run("should normalize remind-at to RFC3339", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateReminderCmd(), "--title", "call back", "--remind-at", "2026-08-20 10:00")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.RemindAt == nil {
			t.Fatal("remind_at = nil, want a value")
		}

		if !strings.HasPrefix(*got.RemindAt, "2026-08-20T10:00:00") {
			t.Errorf("remind_at = %q, want prefix %q", *got.RemindAt, "2026-08-20T10:00:00")
		}
	})

	t.Run("should default the status to pending", func(t *testing.T) {
		useTempVault(t)

		out, err := runCmd(t, newCreateReminderCmd(), "--title", "call back", "--remind-at", "2026-08-20 10:00")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItem(t, out)

		if got.Status == nil {
			t.Fatal("status = nil, want a value")
		}

		if *got.Status != "pending" {
			t.Errorf("status = %q, want %q", *got.Status, "pending")
		}
	})

	t.Run("should fail when the remind-at format is invalid", func(t *testing.T) {
		useTempVault(t)

		_, err := runCmd(t, newCreateReminderCmd(), "--title", "call back", "--remind-at", "amanha")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "invalid --remind-at") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "invalid --remind-at")
		}
	})
}
