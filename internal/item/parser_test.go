package item

import (
	"reflect"
	"strings"
	"testing"
)

const baseFile = `---
id: 2026-08-12-0011-example
title: tech sync
type: note
created: "2026-08-12T00:00:00-03:00"
tags:
    - tech
    - meetings
related:
    - tech-group
---

Lorem ipsum dolor sit amet. Cum ipsum repellendus aut ipsam voluptatem hic rerum tempore qui galisum quos sed nihil exercitationem id galisum commodi eum ducimus enim. In optio repudiandae vel repellendus dolore qui internos odit.`

func TestParse(t *testing.T) {
	t.Run("should return valid parsed item", func(t *testing.T) {
		expected := Item{
			Front: Frontmatter{
				ID:      "2026-08-12-0011-example",
				Title:   "tech sync",
				Type:    "note",
				Created: "2026-08-12T00:00:00-03:00",
				Tags:    []string{"tech", "meetings"},
				Related: []string{"tech-group"},
			},
			Body: "Lorem ipsum dolor sit amet. Cum ipsum repellendus aut ipsam voluptatem hic rerum tempore qui galisum quos sed nihil exercitationem id galisum commodi eum ducimus enim. In optio repudiandae vel repellendus dolore qui internos odit.",
		}

		parsedItem, err := Parse([]byte(baseFile))

		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(expected, parsedItem) {
			t.Errorf("Parse() = %#v, want %#v", parsedItem, expected)
		}
	})

	t.Run("should fail when the file lacks frontmatter ", func(t *testing.T) {
		noDelimiterFile := "id: example\ntitle: no delimiter\n"

		wantMsg := "file lacks frontmatter or is poorly formatted"

		_, err := Parse([]byte(noDelimiterFile))
		if err == nil {
			t.Fatal("Parse() returned nil error, want an error")
		}

		if err.Error() != wantMsg {
			t.Errorf("Parse() error = %q, want %q", err.Error(), wantMsg)
		}
	})

	t.Run("should faile when the frontmatter is poorly formatted", func(t *testing.T) {
		t.Run("with no close delimiter", func(t *testing.T) {
			poorlyFormattedFile := strings.Replace(baseFile, "---\n\n", "", 1)

			wantMsg := "frontmatter is poorly formatted"

			_, err := Parse([]byte(poorlyFormattedFile))
			if err == nil {
				t.Fatal("Parse() returned nil error, want an error")
			}

			if err.Error() != wantMsg {
				t.Errorf("Parse() error = %q, want %q", err.Error(), wantMsg)
			}
		})

		t.Run("should fail when the frontmatter has invalid yaml", func(t *testing.T) {
			invalidYamlFile := "---\nid: [unterminated\n---\nbody\n"

			wantMsg := "frontmatter is poorly formatted"

			_, err := Parse([]byte(invalidYamlFile))
			if err == nil {
				t.Fatal("Parse() returned nil error, want an error")
			}

			wantPrefix := "frontmatter is poorly formatted"
			if !strings.HasPrefix(err.Error(), wantMsg) {
				t.Errorf("Parse() error = %q, want prefix %q", err.Error(), wantPrefix)
			}
		})
	})

}
