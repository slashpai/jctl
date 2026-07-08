package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/jira"
)

type TransitionCmd struct {
	IssueKey string `arg:"" help:"Issue key (e.g. PROJ-123)."`
	Status   string `short:"s" help:"Target status name."`
	List     bool   `help:"List available transitions."`
}

func (t *TransitionCmd) Run() error {
	client, _, err := jiraClient()
	if err != nil {
		return err
	}

	transitions, err := client.GetTransitions(t.IssueKey)
	if err != nil {
		return err
	}

	if t.List || t.Status == "" {
		if len(transitions) == 0 {
			fmt.Printf("No transitions available for %s.\n", t.IssueKey)
			fmt.Println("This may be due to insufficient permissions or the issue is in a terminal state.")
			fmt.Println("Tip: verify your API token has the correct project permissions.")
			return nil
		}
		fmt.Printf("Available transitions for %s:\n\n", t.IssueKey)
		for _, tr := range transitions {
			fmt.Printf("  • %s → %s\n", tr.Name, tr.To.Name)
		}
		if !t.List {
			fmt.Println("\nUse --status to transition, e.g.:")
			fmt.Printf("  jctl issue transition %s -s %q\n", t.IssueKey, transitions[0].To.Name)
		}
		return nil
	}

	var match *jira.Transition
	for i := range transitions {
		if strings.EqualFold(transitions[i].To.Name, t.Status) ||
			strings.EqualFold(transitions[i].Name, t.Status) {
			match = &transitions[i]
			break
		}
	}

	if match == nil {
		available := make([]string, len(transitions))
		for i, tr := range transitions {
			available[i] = tr.To.Name
		}
		return fmt.Errorf("transition %q not found. Available: %s", t.Status, strings.Join(available, ", "))
	}

	if err := client.TransitionIssue(t.IssueKey, match.ID); err != nil {
		return err
	}

	color.Green("✓ %s → %s", t.IssueKey, match.To.Name)
	return nil
}
