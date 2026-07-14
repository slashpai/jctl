package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// CreateIssueInput holds the fields needed to create an issue.
type CreateIssueInput struct {
	ProjectKey  string
	Summary     string
	Description string
	IssueType   string
	Priority    string
	Assignee    string
	Labels      []string
	Components  []string
}

func (c *Client) CreateIssue(in *CreateIssueInput) (*Issue, error) {
	fields := map[string]any{
		"project":   map[string]string{"key": in.ProjectKey},
		"summary":   in.Summary,
		"issuetype": map[string]string{"name": in.IssueType},
	}

	if in.Description != "" {
		fields["description"] = adfText(in.Description)
	}
	if in.Priority != "" {
		fields["priority"] = map[string]string{"name": in.Priority}
	}
	if in.Assignee != "" {
		fields["assignee"] = map[string]string{"accountId": in.Assignee}
	}
	if len(in.Labels) > 0 {
		fields["labels"] = in.Labels
	}
	if len(in.Components) > 0 {
		comps := make([]map[string]string, len(in.Components))
		for i, name := range in.Components {
			comps[i] = map[string]string{"name": name}
		}
		fields["components"] = comps
	}

	body := map[string]any{"fields": fields}

	resp, status, err := c.do("POST", "issue", body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var issue Issue
	if err := json.Unmarshal(resp, &issue); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &issue, nil
}

func (c *Client) GetIssue(issueKey string) (*Issue, error) {
	path := fmt.Sprintf("issue/%s", url.PathEscape(issueKey))
	resp, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var issue Issue
	if err := json.Unmarshal(resp, &issue); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &issue, nil
}

// UpdateIssueInput holds fields to update on an existing issue.
type UpdateIssueInput struct {
	Summary     *string
	Description *string
	Priority    *string
	Assignee    *string
	Labels      []string
	Components  []string
}

func (c *Client) UpdateIssue(issueKey string, in *UpdateIssueInput) error {
	fields := map[string]any{}

	if in.Summary != nil {
		fields["summary"] = *in.Summary
	}
	if in.Description != nil {
		fields["description"] = adfText(*in.Description)
	}
	if in.Priority != nil {
		fields["priority"] = map[string]string{"name": *in.Priority}
	}
	if in.Assignee != nil {
		fields["assignee"] = map[string]string{"accountId": *in.Assignee}
	}
	if in.Labels != nil {
		fields["labels"] = in.Labels
	}
	if in.Components != nil {
		comps := make([]map[string]string, len(in.Components))
		for i, name := range in.Components {
			comps[i] = map[string]string{"name": name}
		}
		fields["components"] = comps
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}

	body := map[string]any{"fields": fields}
	path := fmt.Sprintf("issue/%s", url.PathEscape(issueKey))

	resp, status, err := c.do("PUT", path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return parseError(resp, status)
	}
	return nil
}

func (c *Client) SearchIssues(jql string, maxResults int) (*SearchResult, error) {
	body := map[string]any{
		"jql":        jql,
		"maxResults": maxResults,
		"fields":     []string{"summary", "status", "assignee", "priority", "issuetype", "project", "created", "updated", "labels"},
	}

	resp, status, err := c.do("POST", "search/jql", body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var result SearchResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetTransitions(issueKey string) ([]Transition, error) {
	path := fmt.Sprintf("issue/%s/transitions", url.PathEscape(issueKey))
	resp, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var result TransitionsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return result.Transitions, nil
}

func (c *Client) TransitionIssue(issueKey, transitionID string) error {
	path := fmt.Sprintf("issue/%s/transitions", url.PathEscape(issueKey))
	body := map[string]any{
		"transition": map[string]string{"id": transitionID},
	}

	resp, status, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return parseError(resp, status)
	}
	return nil
}

func (c *Client) AddComment(issueKey, commentBody string) (*Comment, error) {
	path := fmt.Sprintf("issue/%s/comment", url.PathEscape(issueKey))
	body := map[string]any{
		"body": adfText(commentBody),
	}

	resp, status, err := c.do("POST", path, body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var comment Comment
	if err := json.Unmarshal(resp, &comment); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &comment, nil
}

func (c *Client) GetComments(issueKey string) ([]Comment, error) {
	path := fmt.Sprintf("issue/%s/comment?orderBy=created", url.PathEscape(issueKey))
	resp, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, parseError(resp, status)
	}

	var result CommentsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return result.Comments, nil
}

// ExtractADFText extracts plain text from an Atlassian Document Format value.
// ADF is returned by Jira API v3 for description and comment bodies.
func ExtractADFText(v any) string {
	return extractADFTextDepth(v, 0)
}

const maxADFDepth = 20

func extractADFTextDepth(v any, depth int) string {
	if v == nil || depth > maxADFDepth {
		return ""
	}

	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}

	if m["type"] == "text" {
		if text, ok := m["text"].(string); ok {
			return text
		}
		return ""
	}

	content, ok := m["content"].([]any)
	if !ok {
		return ""
	}

	var parts []string
	for _, item := range content {
		node, ok := item.(map[string]any)
		if !ok {
			continue
		}

		nodeType, _ := node["type"].(string)
		inner := extractADFTextDepth(node, depth+1)

		switch nodeType {
		case "paragraph", "heading":
			if inner != "" {
				parts = append(parts, inner)
			}
		case "bulletList", "orderedList":
			if inner != "" {
				parts = append(parts, inner)
			}
		case "listItem":
			if inner != "" {
				parts = append(parts, "  • "+inner)
			}
		default:
			if inner != "" {
				parts = append(parts, inner)
			}
		}
	}

	return strings.Join(parts, "\n")
}

// FindUser searches for a Jira user by email and returns their account ID.
func (c *Client) FindUser(email string) (string, error) {
	path := fmt.Sprintf("user/search?query=%s", url.QueryEscape(email))
	resp, status, err := c.do("GET", path, nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", parseError(resp, status)
	}

	var users []User
	if err := json.Unmarshal(resp, &users); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	if len(users) == 0 {
		return "", fmt.Errorf("no user found for %q", email)
	}
	return users[0].AccountID, nil
}

// ResolveAssignee accepts an email, account ID, or "me" and returns the account ID.
// If the value contains '@' it is treated as an email and looked up.
func (c *Client) ResolveAssignee(value string) (string, error) {
	if value == "me" {
		return c.CurrentUserAccountID()
	}
	if strings.Contains(value, "@") {
		return c.FindUser(value)
	}
	return value, nil
}

// CurrentUserAccountID returns the account ID of the authenticated user.
func (c *Client) CurrentUserAccountID() (string, error) {
	resp, status, err := c.do("GET", "myself", nil)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", parseError(resp, status)
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	if user.AccountID == "" {
		return "", fmt.Errorf("could not determine current user account ID")
	}
	return user.AccountID, nil
}

// adfText wraps plain text in Atlassian Document Format (required by API v3).
func adfText(text string) map[string]any {
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []map[string]any{
			{
				"type": "paragraph",
				"content": []map[string]any{
					{
						"type": "text",
						"text": text,
					},
				},
			},
		},
	}
}

func parseError(body []byte, status int) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		return fmt.Errorf("API error (HTTP %d): %s", status, errResp.Summary())
	}
	const maxErrLen = 256
	msg := string(body)
	if len(msg) > maxErrLen {
		msg = msg[:maxErrLen] + "...(truncated)"
	}
	return fmt.Errorf("API error (HTTP %d): %s", status, msg)
}
