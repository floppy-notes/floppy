package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

func Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating folder for %s: %w", path, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)

	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}
