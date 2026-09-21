package cli

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

var baseRow = database.ItemRow{
	ID:      "2026-08-12-0011-example",
	Title:   "tech sync",
	Type:    "note",
	Created: "2026-08-12T00:00:00-03:00",
	Path:    "/vault/notes/2026/08/2026-08-12-0011-example.md",
}

var baseItem = item.Item{
	Front: item.Frontmatter{
		ID:      "2026-08-12-0011-example",
		Title:   "tech sync",
		Type:    "note",
		Created: "2026-08-12T00:00:00-03:00",
	},
	Path: "/vault/notes/2026/08/2026-08-12-0011-example.md",
}

func TestNullToPtr(t *testing.T) {
	t.Run("should return nil when the value is not valid", func(t *testing.T) {
		got := nullToPtr(sql.NullString{})

		if got != nil {
			t.Errorf("nullToPtr() = %v, want nil", got)
		}
	})

	t.Run("should return the value when it is valid", func(t *testing.T) {
		got := nullToPtr(sql.NullString{String: "open", Valid: true})

		if got == nil {
			t.Fatal("nullToPtr() = nil, want a pointer")
		}

		if *got != "open" {
			t.Errorf("nullToPtr() = %q, want %q", *got, "open")
		}
	})

	t.Run("should return an empty string when it is valid and empty", func(t *testing.T) {
		got := nullToPtr(sql.NullString{String: "", Valid: true})

		if got == nil {
			t.Fatal("nullToPtr() = nil, want a pointer")
		}

		if *got != "" {
			t.Errorf("nullToPtr() = %q, want %q", *got, "")
		}
	})
}

func TestEmptyToPtr(t *testing.T) {
	t.Run("should return nil when the string is empty", func(t *testing.T) {
		got := emptyToPtr("")

		if got != nil {
			t.Errorf("emptyToPtr() = %v, want nil", got)
		}
	})

	t.Run("should return the value when the string is not empty", func(t *testing.T) {
		got := emptyToPtr("open")

		if got == nil {
			t.Fatal("emptyToPtr() = nil, want a pointer")
		}

		if *got != "open" {
			t.Errorf("emptyToPtr() = %q, want %q", *got, "open")
		}
	})
}

func TestWriteItems(t *testing.T) {
	t.Run("should render an empty list as an empty array", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItems(&buf, nil); err != nil {
			t.Fatalf("writeItems() returned unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), `"items": []`) {
			t.Errorf("writeItems() = %q, want it to contain %q", buf.String(), `"items": []`)
		}

		if !strings.Contains(buf.String(), `"count": 0`) {
			t.Errorf("writeItems() = %q, want it to contain %q", buf.String(), `"count": 0`)
		}
	})

	t.Run("should render null for the fields the row does not have", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItems(&buf, []database.ItemRow{baseRow}); err != nil {
			t.Fatalf("writeItems() returned unexpected error: %v", err)
		}

		for _, field := range []string{`"due": null`, `"status": null`, `"remind_at": null`} {
			if !strings.Contains(buf.String(), field) {
				t.Errorf("writeItems() = %q, want it to contain %q", buf.String(), field)
			}
		}
	})

	t.Run("should render the values the row does have", func(t *testing.T) {
		var buf bytes.Buffer

		row := baseRow
		row.Type = "task"
		row.Due = sql.NullString{String: "2026-08-20", Valid: true}
		row.Status = sql.NullString{String: "open", Valid: true}

		if err := writeItems(&buf, []database.ItemRow{row}); err != nil {
			t.Fatalf("writeItems() returned unexpected error: %v", err)
		}

		for _, field := range []string{`"due": "2026-08-20"`, `"status": "open"`, `"count": 1`} {
			if !strings.Contains(buf.String(), field) {
				t.Errorf("writeItems() = %q, want it to contain %q", buf.String(), field)
			}
		}
	})

	t.Run("should count every row", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItems(&buf, []database.ItemRow{baseRow, baseRow, baseRow}); err != nil {
			t.Fatalf("writeItems() returned unexpected error: %v", err)
		}

		if !strings.Contains(buf.String(), `"count": 3`) {
			t.Errorf("writeItems() = %q, want it to contain %q", buf.String(), `"count": 3`)
		}
	})
}

func TestWriteItem(t *testing.T) {
	t.Run("should render null for the fields the item does not have", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItem(&buf, baseItem); err != nil {
			t.Fatalf("writeItem() returned unexpected error: %v", err)
		}

		for _, field := range []string{`"due": null`, `"status": null`, `"remind_at": null`} {
			if !strings.Contains(buf.String(), field) {
				t.Errorf("writeItem() = %q, want it to contain %q", buf.String(), field)
			}
		}
	})

	t.Run("should render the id title and path", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItem(&buf, baseItem); err != nil {
			t.Fatalf("writeItem() returned unexpected error: %v", err)
		}

		for _, field := range []string{`"id": "2026-08-12-0011-example"`, `"title": "tech sync"`, `"path": "/vault/notes`} {
			if !strings.Contains(buf.String(), field) {
				t.Errorf("writeItem() = %q, want it to contain %q", buf.String(), field)
			}
		}
	})

	t.Run("should not wrap a single item in a list", func(t *testing.T) {
		var buf bytes.Buffer

		if err := writeItem(&buf, baseItem); err != nil {
			t.Fatalf("writeItem() returned unexpected error: %v", err)
		}

		if strings.Contains(buf.String(), `"items"`) {
			t.Errorf("writeItem() = %q, want it to not contain %q", buf.String(), `"items"`)
		}
	})
}
