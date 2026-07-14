package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/jira"
)

type UpdateCmd struct {
	IssueKey    string   `arg:"" help:"Issue key (e.g. PROJ-123)."`
	Summary     *string  `short:"s" help:"New summary."`
	Description *string  `short:"d" help:"New description."`
	Priority    *string  `help:"New priority."`
	Assignee    *string  `help:"New assignee (email, account ID, or 'me' for yourself)."`
	Label       []string `short:"l" help:"Replace labels."`
	Component   []string `help:"Replace components."`
	Comment     string   `short:"c" help:"Add a comment."`
}

func (u *UpdateCmd) Run() error {
	client, cfg, err := jiraClient()
	if err != nil {
		return err
	}

	input := &jira.UpdateIssueInput{}
	hasFieldUpdate := false

	if u.Summary != nil {
		input.Summary = u.Summary
		hasFieldUpdate = true
	}
	if u.Description != nil {
		input.Description = u.Description
		hasFieldUpdate = true
	}
	if u.Priority != nil {
		input.Priority = u.Priority
		hasFieldUpdate = true
	}
	if u.Assignee != nil {
		resolved, err := client.ResolveAssignee(*u.Assignee)
		if err != nil {
			return fmt.Errorf("resolving assignee: %w", err)
		}
		input.Assignee = &resolved
		hasFieldUpdate = true
	}
	if len(u.Label) > 0 {
		input.Labels = u.Label
		hasFieldUpdate = true
	}
	if len(u.Component) > 0 {
		input.Components = u.Component
		hasFieldUpdate = true
	}

	if hasFieldUpdate {
		if err := client.UpdateIssue(u.IssueKey, input); err != nil {
			return err
		}
		color.Green("✓ Updated %s", u.IssueKey)
	}

	var commentID string
	if u.Comment != "" {
		comment, err := client.AddComment(u.IssueKey, u.Comment)
		if err != nil {
			return fmt.Errorf("adding comment: %w", err)
		}
		commentID = comment.ID
		color.Green("✓ Comment added to %s", u.IssueKey)
	}

	if !hasFieldUpdate && u.Comment == "" {
		return fmt.Errorf("specify at least one field to update or --comment")
	}

	if commentID != "" {
		fmt.Printf("  URL: %s/browse/%s?focusedCommentId=%s\n", cfg.BaseURL, u.IssueKey, commentID)
	} else {
		fmt.Printf("  URL: %s/browse/%s\n", cfg.BaseURL, u.IssueKey)
	}
	return nil
}
