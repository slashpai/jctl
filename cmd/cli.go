package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/config"
	"github.com/slashpai/jctl/internal/jira"
)

// CLI defines the top-level command structure for jctl.
type CLI struct {
	Configure ConfigureCmd `cmd:"" help:"Set up Jira connection credentials."`
	Issue     IssueCmd     `cmd:"" help:"Manage Jira issues."`
}

// IssueCmd groups all issue subcommands.
type IssueCmd struct {
	Create     CreateCmd     `cmd:"" help:"Create a new Jira issue."`
	View       ViewCmd       `cmd:"" help:"View details of a Jira issue."`
	Update     UpdateCmd     `cmd:"" help:"Update fields on a Jira issue."`
	List       ListCmd       `cmd:"" aliases:"ls" help:"List Jira issues."`
	Transition TransitionCmd `cmd:"" aliases:"move" help:"Transition an issue to a new status."`
}

func printField(label, value string) {
	if value == "" {
		value = "—"
	}
	labelColor := color.New(color.FgYellow)
	labelColor.Printf("  %-12s", label)
	color.New().Printf(" %s\n", value)
}

func fieldStr[T any](v *T, fn func(*T) string) string {
	if v == nil {
		return ""
	}
	return fn(v)
}

// hyperlink wraps text in an OSC 8 terminal hyperlink escape sequence.
// Falls back to plain text when color/escape output is disabled (e.g. piped output).
func hyperlink(url, text string) string {
	if color.NoColor {
		return text
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// jiraClient is a function variable so tests can swap it with a mock.
var jiraClient = defaultJiraClient

func defaultJiraClient() (*jira.Client, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	return jira.NewClient(cfg), cfg, nil
}
