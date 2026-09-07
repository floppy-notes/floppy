package item

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func Parse(raw []byte) (Item, error) {
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))

	if !bytes.HasPrefix(raw, []byte(delim+"\n")) {
		return Item{}, errors.New("file lacks frontmatter or is poorly formatted")
	}

	raw = raw[len(delim)+1:]

	frontmatterEnd, bodyStart, ok := findCloseDelim(raw)

	if !ok {
		return Item{}, errors.New("frontmatter is poorly formatted")
	}

	rawFrontmatter := raw[:frontmatterEnd]
	rawBody := raw[bodyStart:]

	var frontmatter Frontmatter
	if err := yaml.Unmarshal(rawFrontmatter, &frontmatter); err != nil {
		return Item{}, fmt.Errorf("%w: %v", errors.New("frontmatter is poorly formatted"), err)
	}

	return Item{
		Front: frontmatter,
		Body:  strings.TrimLeft(string(rawBody), "\n"),
	}, nil
}

func findCloseDelim(rest []byte) (int, int, bool) {
	offset := 0

	for line := range bytes.SplitSeq(rest, []byte("\n")) {
		if string(bytes.TrimSpace(line)) == delim {
			yamlEnd := offset
			bodyStart := min(offset+len(line)+1, len(rest))
			return yamlEnd, bodyStart, true
		}

		offset += len(line) + 1
	}

	return 0, 0, false
}
