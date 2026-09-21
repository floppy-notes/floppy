package cli

import (
	"strings"
	"testing"
)

func TestIsReleaseVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"tagged release", "v0.1.0", true},
		{"tagged release with patch", "v1.24.3", true},
		{"prerelease tag", "v1.0.0-rc1", true},
		{"empty", "", false},
		{"devel", "(devel)", false},
		{"pseudo version", "v0.0.0-20230129092748-24d4a6f8daec", false},
		{"pseudo version after a tag", "v1.2.3-0.20230129092748-24d4a6f8daec", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReleaseVersion(tt.in)

			if got != tt.want {
				t.Errorf("isReleaseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	t.Run("should report the go version and platform", func(t *testing.T) {
		got := versionString()

		if !strings.Contains(got, "go1.") {
			t.Errorf("versionString() = %q, want it to contain %q", got, "go1.")
		}

		if len(strings.Fields(got)) < 3 {
			t.Errorf("versionString() = %q, want at least 3 space separated fields", got)
		}
	})

	t.Run("should fall back to dev when no version is set", func(t *testing.T) {
		original := version
		t.Cleanup(func() { version = original })

		version = ""

		got := versionString()

		if !strings.HasPrefix(got, "dev") {
			t.Errorf("versionString() = %q, want prefix %q", got, "dev")
		}
	})

	t.Run("should use the version set at build time", func(t *testing.T) {
		original := version
		t.Cleanup(func() { version = original })

		version = "v9.9.9"

		got := versionString()

		if !strings.HasPrefix(got, "v9.9.9") {
			t.Errorf("versionString() = %q, want prefix %q", got, "v9.9.9")
		}
	})
}
