package completion

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

type testCLI struct {
	Configure struct{} `cmd:""`
	Issue     struct {
		Create struct {
			Project string `short:"p"`
			Summary string `short:"s"`
			Type    string `short:"t"`
		} `cmd:""`
		View struct {
			IssueKey string `arg:""`
			Web      bool   `short:"w"`
			Comments bool   `short:"c"`
		} `cmd:""`
		Update     struct{} `cmd:""`
		List       struct{} `cmd:"" aliases:"ls"`
		Transition struct{} `cmd:"" aliases:"move"`
	} `cmd:""`
	Completion struct {
		Shell string `arg:"" optional:"" enum:"bash,zsh,fish," default:""`
	} `cmd:""`
}

func testParser(t *testing.T) *kong.Kong {
	t.Helper()
	var cli testCLI
	parser, err := kong.New(&cli, kong.Name("jctl"))
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	return parser
}

func TestPredictTopLevelCommands(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl ", 5)
	want := []string{"configure", "issue", "completion"}
	assertContainsAll(t, got, want)
}

func TestPredictIssueSubcommands(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl issue ", 11)
	want := []string{"create", "view", "update", "list", "ls", "transition", "move"}
	assertContainsAll(t, got, want)
}

func TestPredictIssuePrefix(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl i", 6)
	if len(got) != 1 || got[0] != "issue" {
		t.Fatalf("expected [issue], got %v", got)
	}
}

func TestPredictCreateFlags(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl issue create -", 19)
	want := []string{"--project", "-p", "--summary", "-s", "--type", "-t"}
	assertContainsAll(t, got, want)
}

func TestPredictCreateFlagsAfterCommand(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl issue create ", 18)
	want := []string{"--project", "-p", "--summary", "-s", "--type", "-t"}
	assertContainsAll(t, got, want)
}

func TestPredictCreateFlagsAfterFlagValue(t *testing.T) {
	parser := testParser(t)
	line := "jctl issue create -p PROJ "
	got := Predict(parser.Model.Node, line, len(line))
	want := []string{"--project", "-p", "--summary", "-s", "--type", "-t"}
	assertContainsAll(t, got, want)
}

func TestPredictViewFlagsAfterIssueKey(t *testing.T) {
	parser := testParser(t)
	line := "jctl issue view PROJ-123 "
	got := Predict(parser.Model.Node, line, len(line))
	want := []string{"--web", "-w", "--comments", "-c"}
	assertContainsAll(t, got, want)
}

func TestPredictViewFlagsAfterSubcommand(t *testing.T) {
	parser := testParser(t)
	line := "jctl issue view "
	got := Predict(parser.Model.Node, line, len(line))
	want := []string{"--web", "-w", "--comments", "-c"}
	assertContainsAll(t, got, want)
}

func TestPredictCompletionShellEnum(t *testing.T) {
	parser := testParser(t)
	got := Predict(parser.Model.Node, "jctl completion ", 16)
	want := []string{"bash", "zsh", "fish"}
	assertContainsAll(t, got, want)
}

func TestShellScriptBash(t *testing.T) {
	script, err := ShellScript("bash", "jctl", "/usr/local/bin/jctl")
	if err != nil {
		t.Fatalf("ShellScript: %v", err)
	}
	if !strings.Contains(script, "-C /usr/local/bin/jctl jctl") {
		t.Fatalf("unexpected bash script: %q", script)
	}
}

func TestShellScriptUnsupported(t *testing.T) {
	_, err := ShellScript("powershell", "jctl", "/usr/local/bin/jctl")
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

func assertContainsAll(t *testing.T, got, want []string) {
	t.Helper()
	seen := make(map[string]struct{}, len(got))
	for _, item := range got {
		seen[item] = struct{}{}
	}
	for _, item := range want {
		if _, ok := seen[item]; !ok {
			t.Fatalf("missing %q in %v", item, got)
		}
	}
}
