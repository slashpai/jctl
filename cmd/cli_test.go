package cmd

import (
	"testing"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/jira"
)

func TestFieldStr(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		got := fieldStr[jira.Status](nil, func(s *jira.Status) string { return s.Name })
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		s := &jira.Status{Name: "Open"}
		got := fieldStr(s, func(s *jira.Status) string { return s.Name })
		if got != "Open" {
			t.Errorf("expected 'Open', got %q", got)
		}
	})
}

func TestHyperlink(t *testing.T) {
	t.Run("with color enabled", func(t *testing.T) {
		origNoColor := color.NoColor
		color.NoColor = false
		defer func() { color.NoColor = origNoColor }()

		got := hyperlink("https://example.com", "click me")
		if got == "click me" {
			t.Error("expected OSC 8 escape sequence, got plain text")
		}
		if len(got) <= len("click me") {
			t.Error("expected output longer than plain text")
		}
	})

	t.Run("with color disabled", func(t *testing.T) {
		origNoColor := color.NoColor
		color.NoColor = true
		defer func() { color.NoColor = origNoColor }()

		got := hyperlink("https://example.com", "click me")
		if got != "click me" {
			t.Errorf("expected plain text, got %q", got)
		}
	})
}
