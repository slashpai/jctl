package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestCompletionCmdCode(t *testing.T) {
	var cli CLI
	parser, err := kong.New(&cli, kong.Name("jctl"))
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}

	var stdout bytes.Buffer
	parser.Stdout = &stdout

	ctx, err := parser.Parse([]string{"completion", "-c", "bash"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ctx.Stdout = &stdout

	if err := ctx.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(stdout.String(), "complete -o default -o bashdefault -C") || !strings.Contains(stdout.String(), " jctl") {
		t.Fatalf("expected completion script, got %q", stdout.String())
	}
}
