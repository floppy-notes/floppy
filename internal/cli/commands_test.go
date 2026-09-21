package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetGlobalFlags(t *testing.T) {
	t.Helper()

	originalList := listFlags
	originalLimit := searchLimit
	originalRebuild := indexRebuild

	t.Cleanup(func() {
		listFlags = originalList
		searchLimit = originalLimit
		indexRebuild = originalRebuild
	})
}

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer

	rootCmd.SetArgs(append([]string{"--vault", vaultPath}, args...))
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	err := rootCmd.Execute()

	return out.String(), err
}

func decodeItems(t *testing.T, raw string) itemsResult {
	t.Helper()

	var result itemsResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("Unmarshal() returned unexpected error: %v, raw = %q", err, raw)
	}

	return result
}

func TestListCmd(t *testing.T) {
	t.Run("should list every item", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")
		seedTask(t, "write tests")

		out, err := runRoot(t, "list", "--limit", "10")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 2 {
			t.Errorf("count = %d, want %d", got.Count, 2)
		}
	})

	t.Run("should filter by type", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")
		seedTask(t, "write tests")

		out, err := runRoot(t, "list", "--type", "task", "--limit", "10")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 1 {
			t.Fatalf("count = %d, want %d", got.Count, 1)
		}

		if got.Items[0].Type != "task" {
			t.Errorf("type = %q, want %q", got.Items[0].Type, "task")
		}
	})

	t.Run("should respect the limit", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "first")
		seedNote(t, "second")

		out, err := runRoot(t, "list", "--limit", "1")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 1 {
			t.Errorf("count = %d, want %d", got.Count, 1)
		}
	})

	t.Run("should render an empty vault", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		out, err := runRoot(t, "list", "--limit", "10")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 0 {
			t.Errorf("count = %d, want %d", got.Count, 0)
		}
	})

	t.Run("should fail when the flags are invalid", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		_, err := runRoot(t, "list", "--type", "event", "--limit", "10")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "validating flags") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "validating flags")
		}
	})

	t.Run("should fail when the status does not match the type", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		_, err := runRoot(t, "list", "--type", "task", "--status", "pending", "--limit", "10")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "invalid task status") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "invalid task status")
		}
	})
}

func TestSearchCmd(t *testing.T) {
	t.Run("should find an item by its title", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "deploy pipeline")
		seedNote(t, "unrelated note")

		out, err := runRoot(t, "search", "pipeline")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 1 {
			t.Fatalf("count = %d, want %d", got.Count, 1)
		}

		if got.Items[0].Title != "deploy pipeline" {
			t.Errorf("title = %q, want %q", got.Items[0].Title, "deploy pipeline")
		}
	})

	t.Run("should join several terms into one query", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "deploy pipeline")

		out, err := runRoot(t, "search", "deploy", "pipeline")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 1 {
			t.Errorf("count = %d, want %d", got.Count, 1)
		}
	})

	t.Run("should render nothing when there is no match", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")

		out, err := runRoot(t, "search", "kubernetes")

		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		got := decodeItems(t, out)

		if got.Count != 0 {
			t.Errorf("count = %d, want %d", got.Count, 0)
		}
	})

	t.Run("should fail when the query syntax is invalid", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")

		_, err := runRoot(t, "search", `"unbalanced`)

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.Contains(err.Error(), "check for unbalanced quotes") {
			t.Errorf("Execute() error = %q, want it to contain %q", err.Error(), "check for unbalanced quotes")
		}
	})

	t.Run("should require a query", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		_, err := runRoot(t, "search")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}
	})
}

func TestIndexCmd(t *testing.T) {
	t.Run("should rebuild the index from the vault", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")

		if _, err := runRoot(t, "index", "--rebuild"); err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		out, err := runRoot(t, "list", "--limit", "10")
		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if got := decodeItems(t, out); got.Count != 1 {
			t.Errorf("count = %d, want %d", got.Count, 1)
		}
	})

	t.Run("should refuse incremental indexing", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		_, err := runRoot(t, "index")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		wantMsg := "incremental indexing not implemented yet, use --rebuild"
		if err.Error() != wantMsg {
			t.Errorf("Execute() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should drop the items that left the vault", func(t *testing.T) {
		useTempVault(t)
		resetGlobalFlags(t)

		seedNote(t, "tech sync")

		if err := removeVaultNotes(t); err != nil {
			t.Fatalf("removing the notes folder returned unexpected error: %v", err)
		}

		if _, err := runRoot(t, "index", "--rebuild"); err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		out, err := runRoot(t, "list", "--limit", "10")
		if err != nil {
			t.Fatalf("Execute() returned unexpected error: %v", err)
		}

		if got := decodeItems(t, out); got.Count != 0 {
			t.Errorf("count = %d, want %d", got.Count, 0)
		}
	})
}

// removeVaultNotes deletes the notes folder of the current vault, simulating a file
// the user removed by hand.
func removeVaultNotes(t *testing.T) error {
	t.Helper()

	return os.RemoveAll(filepath.Join(vaultPath, "notes"))
}

func TestRunCreateErrors(t *testing.T) {
	t.Run("should fail when the database cannot be opened", func(t *testing.T) {
		resetGlobalFlags(t)

		original := vaultPath
		t.Cleanup(func() { vaultPath = original })

		blocker := filepath.Join(t.TempDir(), "a-file")
		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}
		vaultPath = filepath.Join(blocker, "nested")

		_, err := runCmd(t, newCreateNoteCmd(), "--title", "tech sync")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "opening database") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "opening database")
		}
	})

	t.Run("should fail when the item cannot be created", func(t *testing.T) {
		dir := useTempVault(t)
		resetGlobalFlags(t)

		blocker := filepath.Join(dir, "notes")
		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}

		_, err := runCmd(t, newCreateNoteCmd(), "--title", "tech sync")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "building id") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "building id")
		}
	})
}

func TestRunUpdateErrors(t *testing.T) {
	t.Run("should fail when the database cannot be opened", func(t *testing.T) {
		resetGlobalFlags(t)

		original := vaultPath
		t.Cleanup(func() { vaultPath = original })

		blocker := filepath.Join(t.TempDir(), "a-file")
		if err := os.WriteFile(blocker, []byte("not a folder"), 0o644); err != nil {
			t.Fatalf("WriteFile() returned unexpected error: %v", err)
		}
		vaultPath = filepath.Join(blocker, "nested")

		_, err := runCmd(t, newUpdateNoteCmd(), "--id", "any", "--title", "x")

		if err == nil {
			t.Fatal("Execute() returned nil error, want an error")
		}

		if !strings.HasPrefix(err.Error(), "opening database") {
			t.Errorf("Execute() error = %q, want prefix %q", err.Error(), "opening database")
		}
	})
}
