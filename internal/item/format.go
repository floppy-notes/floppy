package item

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Format(item Item) ([]byte, error) {
	if err := item.Front.Validate(); err != nil {
		return nil, fmt.Errorf("invalid item: %w", err)
	}

	parsedFrontmatter, err := yaml.Marshal(item.Front)

	if err != nil {
		return nil, fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(delim + "\n")
	buf.Write(parsedFrontmatter)
	buf.WriteString(delim + "\n\n")
	buf.WriteString(item.Body)

	return buf.Bytes(), nil
}
