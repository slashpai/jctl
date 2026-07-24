package version

import "testing"

func TestResolveVersion_ReleaseTag(t *testing.T) {
	got := resolveVersion("v1.2.3", "abc1234")
	if got != "v1.2.3" {
		t.Fatalf("expected release tag, got %q", got)
	}
}

func TestResolveVersion_CommitWhenNoTag(t *testing.T) {
	got := resolveVersion("", "04f5a35")
	if got != "04f5a35" {
		t.Fatalf("expected commit hash, got %q", got)
	}
}

func TestResolveVersion_DevFallback(t *testing.T) {
	got := resolveVersion("", "")
	if got == "" {
		t.Fatal("expected non-empty version string")
	}
}

func TestString_ReturnsNonEmpty(t *testing.T) {
	got := String()
	if got == "" {
		t.Fatal("expected non-empty version string")
	}
}
