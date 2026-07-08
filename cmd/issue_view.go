package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/jira"
)

type ViewCmd struct {
	IssueKey string `arg:"" help:"Issue key (e.g. PROJ-123)."`
	Web      bool   `short:"w" help:"Open issue in browser."`
	Comments bool   `short:"c" help:"Show issue comments."`
}

func (v *ViewCmd) Run() error {
	client, cfg, err := jiraClient()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(v.IssueKey)
	if err != nil {
		return err
	}

	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	issueURL := fmt.Sprintf("%s/browse/%s", cfg.BaseURL, issue.Key)
	cyan.Printf("%s  ", hyperlink(issueURL, issue.Key))
	bold.Println(issue.Fields.Summary)
	fmt.Println(strings.Repeat("─", 50))

	printField("Type", fieldStr(issue.Fields.IssueType, func(t *jira.IssueType) string { return t.Name }))
	printField("Status", fieldStr(issue.Fields.Status, func(s *jira.Status) string { return s.Name }))
	printField("Priority", fieldStr(issue.Fields.Priority, func(p *jira.Priority) string { return p.Name }))
	printField("Assignee", fieldStr(issue.Fields.Assignee, func(u *jira.User) string { return u.DisplayName }))
	printField("Reporter", fieldStr(issue.Fields.Reporter, func(u *jira.User) string { return u.DisplayName }))

	if len(issue.Fields.Labels) > 0 {
		printField("Labels", strings.Join(issue.Fields.Labels, ", "))
	}
	if issue.Fields.Created != "" {
		printField("Created", issue.Fields.Created)
	}
	if issue.Fields.Updated != "" {
		printField("Updated", issue.Fields.Updated)
	}

	desc := jira.ExtractADFText(issue.Fields.Description)
	if desc != "" {
		fmt.Println()
		descHeader := color.New(color.FgYellow)
		descHeader.Println("  Description")
		fmt.Println("  " + strings.Repeat("─", 40))
		for _, line := range strings.Split(desc, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}

	if v.Comments {
		comments, err := client.GetComments(v.IssueKey)
		if err != nil {
			return fmt.Errorf("fetching comments: %w", err)
		}

		fmt.Println()
		commentHeader := color.New(color.FgYellow)
		commentHeader.Printf("  Comments (%d)\n", len(comments))
		fmt.Println("  " + strings.Repeat("─", 40))

		if len(comments) == 0 {
			fmt.Println("  No comments")
		}
		for i, c := range comments {
			author := "Unknown"
			if c.Author != nil {
				author = c.Author.DisplayName
			}
			authorColor := color.New(color.FgCyan)
			fmt.Print("  ")
			authorColor.Printf("%s", author)
			dimColor := color.New(color.Faint)
			dimColor.Printf("  %s\n", c.Created)

			body := jira.ExtractADFText(c.Body)
			if body != "" {
				for _, line := range strings.Split(body, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}
			if i < len(comments)-1 {
				fmt.Println()
			}
		}
	}

	fmt.Printf("\n  URL: %s/browse/%s\n", cfg.BaseURL, issue.Key)
	return nil
}
