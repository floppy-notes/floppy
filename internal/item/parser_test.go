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

	t.Run("should keep an empty body", func(t *testing.T) {
		emptyBodyFile := "---\nid: id\ntitle: t\ntype: note\ncreated: \"2026-08-12T00:00:00-03:00\"\n---\n\n"

		parsedItem, err := Parse([]byte(emptyBodyFile))

		if err != nil {
			t.Fatalf("Parse() returned unexpected error: %v", err)
		}

		if parsedItem.Body != "" {
			t.Errorf("Parse() body = %q, want %q", parsedItem.Body, "")
		}
	})
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantPrefix string
	}{
		{
			"no frontmatter delimiter",
			"id: example\ntitle: no delimiter\n",
			"file lacks frontmatter or is poorly formatted",
		},
		{
			"no close delimiter",
			strings.Replace(baseFile, "---\n\n", "", 1),
			"frontmatter is poorly formatted",
		},
		{
			"invalid yaml",
			"---\nid: [unterminated\n---\nbody\n",
			"frontmatter is poorly formatted",
		},
		{
			"empty file",
			"",
			"file lacks frontmatter or is poorly formatted",
		},
	}

	for _, tt := range tests {
		t.Run("should fail when the file has "+tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.in))

			if err == nil {
				t.Fatal("Parse() returned nil error, want an error")
			}

			if !strings.HasPrefix(err.Error(), tt.wantPrefix) {
				t.Errorf("Parse() error = %q, want prefix %q", err.Error(), tt.wantPrefix)
			}
		})
	}
}
