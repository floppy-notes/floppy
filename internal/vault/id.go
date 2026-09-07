package vault

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func BuildId(folderName string, name string, createdTime time.Time) (string, error) {
	formatedTime := createdTime.Format("2006-01-02")

	splitString := strings.Split(formatedTime, "-")

	targetFolder := filepath.Join(folderName, splitString[0], splitString[1])

	nextId, err := nextIdSequence(targetFolder)

	if err != nil {
		return "", fmt.Errorf("building id: %w", err)
	}

	builtId := fmt.Sprintf("%s-%04d-%s", formatedTime, nextId, slugfy(name))

	return builtId, nil
}

func nextIdSequence(targetFolder string) (int, error) {
	items, _, err := Walk(targetFolder)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 1, nil
		}
		return 0, fmt.Errorf("walking %s: %w", targetFolder, err)
	}

	lastestSeq := 0

	for _, i := range items {
		parts := strings.Split(i.Front.ID, "-")

		if len(parts) < 4 {
			return 0, fmt.Errorf("invalid id format in %s", i.Path)
		}

		seq, err := strconv.Atoi(parts[3])

		if err != nil {
			return 0, fmt.Errorf("invalid sequence id. It should be an integer: %w", err)
		}

		lastestSeq = int(max(seq, lastestSeq))

	}

	return lastestSeq + 1, nil
}

func slugfy(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	ascii, _, _ := transform.String(t, s)

	slug := nonAlnum.ReplaceAllString(strings.ToLower(ascii), "-")
	return strings.Trim(slug, "-")
}
