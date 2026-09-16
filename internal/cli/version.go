package cli

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
)

var version string

var pseudoVersion = regexp.MustCompile(`\d{14}-[0-9a-f]{12}`)

func versionString() string {
	v := version
	revision := ""
	dirty := false

	if info, ok := debug.ReadBuildInfo(); ok {
		if v == "" && isReleaseVersion(info.Main.Version) {
			v = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
	}

	if v == "" {
		v = "dev"
	}

	if len(revision) > 7 {
		revision = revision[:7]
	}

	if revision != "" {
		if dirty {
			revision += "-dirty"
		}
		v = fmt.Sprintf("%s (%s)", v, revision)
	}

	return fmt.Sprintf("%s %s %s/%s", v, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func isReleaseVersion(v string) bool {
	if v == "" || v == "(devel)" {
		return false
	}
	return !pseudoVersion.MatchString(v)
}
