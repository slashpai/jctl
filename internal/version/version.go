package version

import (
	"runtime/debug"
	"strings"
)

// Set via -ldflags at build time.
var (
	version string
	commit  string
)

// String returns the version to display.
// For tagged releases it returns the release tag; otherwise it returns the
// latest git commit (short hash), falling back to "dev" when unavailable.
func String() string {
	return resolveVersion(version, commit)
}

func resolveVersion(ver, cmt string) string {
	if v := strings.TrimSpace(ver); v != "" && v != "dev" {
		return v
	}
	if c := resolveCommit(cmt); c != "" {
		return c
	}
	return "dev"
}

func resolveCommit(cmt string) string {
	if c := strings.TrimSpace(cmt); c != "" && c != "none" {
		return c
	}
	return vcsRevision()
}

func vcsRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	var revision string
	var dirty bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}
	if revision == "" {
		return ""
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if dirty {
		return revision + "-dirty"
	}
	return revision
}
