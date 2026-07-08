package jira

// Issue represents a Jira issue with commonly used fields.
type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary     string      `json:"summary"`
	Description any         `json:"description,omitempty"`
	Status      *Status     `json:"status,omitempty"`
	Assignee    *User       `json:"assignee,omitempty"`
	Reporter    *User       `json:"reporter,omitempty"`
	Priority    *Priority   `json:"priority,omitempty"`
	IssueType   *IssueType  `json:"issuetype,omitempty"`
	Project     *Project    `json:"project,omitempty"`
	Labels      []string    `json:"labels,omitempty"`
	Created     string      `json:"created,omitempty"`
	Updated     string      `json:"updated,omitempty"`
	Components  []Component `json:"components,omitempty"`
}

type Status struct {
	Name string `json:"name"`
}

type User struct {
	DisplayName string `json:"displayName"`
	AccountID   string `json:"accountId"`
	Email       string `json:"emailAddress,omitempty"`
}

type Priority struct {
	Name string `json:"name"`
}

type IssueType struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

type Project struct {
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

type Component struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   Status `json:"to"`
}

// SearchResult is the response from the issue search endpoint (search/jql).
type SearchResult struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

// Comment represents a single Jira issue comment.
type Comment struct {
	ID      string `json:"id"`
	Author  *User  `json:"author,omitempty"`
	Body    any    `json:"body,omitempty"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// CommentsResponse wraps the paginated list of comments.
type CommentsResponse struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Comments   []Comment `json:"comments"`
}

// TransitionsResponse wraps the list of available transitions.
type TransitionsResponse struct {
	Transitions []Transition `json:"transitions"`
}

type ErrorResponse struct {
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

func (e *ErrorResponse) Summary() string {
	msg := ""
	for _, m := range e.ErrorMessages {
		msg += m + "; "
	}
	for k, v := range e.Errors {
		msg += k + ": " + v + "; "
	}
	if msg == "" {
		return "unknown error"
	}
	return msg
}
