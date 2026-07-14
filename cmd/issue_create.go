package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/jira"
)

type CreateCmd struct {
	Project     string   `short:"p" required:"" help:"Project key."`
	Summary     string   `short:"s" required:"" help:"Issue summary."`
	Description string   `short:"d" help:"Issue description."`
	Type        string   `short:"t" default:"Task" help:"Issue type (e.g. Bug, Story, Task)."`
	Priority    string   `help:"Priority (e.g. High, Medium, Low)."`
	Assignee    string   `default:"me" help:"Assignee (defaults to 'me'; use 'none' to leave unassigned, or an email/account ID)."`
	Label       []string `short:"l" help:"Labels (can be repeated)."`
	Component   []string `help:"Components (can be repeated)."`
}

func (c *CreateCmd) Run() error {
	client, cfg, err := jiraClient()
	if err != nil {
		return err
	}

	assignee := c.Assignee
	if assignee == "" {
		assignee = "me"
	}
	if assignee == "none" {
		assignee = ""
	} else {
		assignee, err = client.ResolveAssignee(assignee)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
	}

	issue, err := client.CreateIssue(&jira.CreateIssueInput{
		ProjectKey:  c.Project,
		Summary:     c.Summary,
		Description: c.Description,
		IssueType:   c.Type,
		Priority:    c.Priority,
		Assignee:    assignee,
		Labels:      c.Label,
		Components:  c.Component,
	})
	if err != nil {
		return err
	}

	color.Green("✓ Created %s", issue.Key)
	fmt.Printf("  URL: %s/browse/%s\n", cfg.BaseURL, issue.Key)
	return nil
}
