package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type ListCmd struct {
	Project  string `short:"p" help:"Filter by project key."`
	Assignee string `short:"a" help:"Filter by assignee (use 'me' for current user)."`
	Status   string `help:"Filter by status."`
	Type     string `short:"t" help:"Filter by issue type."`
	JQL      string `help:"Raw JQL query (overrides other filters)."`
	Max      int    `short:"n" default:"25" help:"Maximum results to return."`
}

func (l *ListCmd) Run() error {
	client, cfg, err := jiraClient()
	if err != nil {
		return err
	}

	jqlQuery := l.JQL
	if jqlQuery == "" {
		jqlQuery = buildJQL(l)
	}
	if jqlQuery == "" {
		return fmt.Errorf("specify at least --project or --jql")
	}

	result, err := client.SearchIssues(jqlQuery, l.Max)
	if err != nil {
		return err
	}

	if len(result.Issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	type row struct {
		key, typ, status, priority, assignee, summary string
	}

	rows := make([]row, len(result.Issues))
	colW := [6]int{3, 4, 6, 8, 8, 7} // min widths from header labels

	for i, issue := range result.Issues {
		r := row{key: issue.Key}
		if issue.Fields.IssueType != nil {
			r.typ = issue.Fields.IssueType.Name
		}
		if issue.Fields.Status != nil {
			r.status = issue.Fields.Status.Name
		}
		if issue.Fields.Priority != nil {
			r.priority = issue.Fields.Priority.Name
		}
		r.assignee = "Unassigned"
		if issue.Fields.Assignee != nil {
			r.assignee = issue.Fields.Assignee.DisplayName
		}
		r.summary = issue.Fields.Summary
		if len(r.summary) > 60 {
			r.summary = r.summary[:57] + "..."
		}
		rows[i] = r

		for j, v := range []string{r.key, r.typ, r.status, r.priority, r.assignee} {
			if len(v) > colW[j] {
				colW[j] = len(v)
			}
		}
	}

	fmtStr := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n",
		colW[0], colW[1], colW[2], colW[3], colW[4])
	sepStr := fmt.Sprintf("  %s──%s──%s──%s──%s──%s",
		strings.Repeat("─", colW[0]),
		strings.Repeat("─", colW[1]),
		strings.Repeat("─", colW[2]),
		strings.Repeat("─", colW[3]),
		strings.Repeat("─", colW[4]),
		strings.Repeat("─", 20))

	fmt.Printf(fmtStr, "KEY", "TYPE", "STATUS", "PRIORITY", "ASSIGNEE", "SUMMARY")
	fmt.Println(sepStr)

	cyan := color.New(color.FgCyan)
	for _, r := range rows {
		issueURL := fmt.Sprintf("%s/browse/%s", cfg.BaseURL, r.key)
		link := hyperlink(issueURL, fmt.Sprintf("%-*s", colW[0], r.key))
		fmt.Print("  ")
		cyan.Print(link)
		fmt.Printf("  %-*s  %-*s  %-*s  %-*s  %s\n",
			colW[1], r.typ,
			colW[2], r.status,
			colW[3], r.priority,
			colW[4], r.assignee,
			r.summary)
	}

	fmt.Printf("\n  %d issue(s) shown", len(rows))
	if result.NextPageToken != "" {
		fmt.Printf(" · more results available, use -n to increase the limit")
	}
	fmt.Println()

	return nil
}

func buildJQL(l *ListCmd) string {
	var clauses []string

	if l.Project != "" {
		clauses = append(clauses, fmt.Sprintf("project = %q", l.Project))
	}
	if l.Assignee != "" {
		if l.Assignee == "me" {
			clauses = append(clauses, "assignee = currentUser()")
		} else {
			clauses = append(clauses, fmt.Sprintf("assignee = %q", l.Assignee))
		}
	}
	if l.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = %q", l.Status))
	}
	if l.Type != "" {
		clauses = append(clauses, fmt.Sprintf("issuetype = %q", l.Type))
	}

	if len(clauses) == 0 {
		return ""
	}

	query := clauses[0]
	for _, c := range clauses[1:] {
		query += " AND " + c
	}
	return query + " ORDER BY updated DESC"
}
