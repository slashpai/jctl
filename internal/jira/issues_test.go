package jira

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slashpai/jctl/internal/config"
)

func testClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{
		BaseURL: srv.URL,
		Email:   "test@example.com",
		Token:   "test-token",
	}
	return NewClient(cfg), srv
}

func TestCreateIssue(t *testing.T) {
	t.Run("minimal fields", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/issue" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			assertBasicAuth(t, r, "test@example.com", "test-token")
			assertContentType(t, r)

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)

			if fields["summary"] != "Test issue" {
				t.Errorf("expected summary 'Test issue', got %v", fields["summary"])
			}
			proj := fields["project"].(map[string]any)
			if proj["key"] != "PROJ" {
				t.Errorf("expected project key 'PROJ', got %v", proj["key"])
			}
			itype := fields["issuetype"].(map[string]any)
			if itype["name"] != "Bug" {
				t.Errorf("expected issue type 'Bug', got %v", itype["name"])
			}
			if _, ok := fields["description"]; ok {
				t.Error("description should not be set for empty input")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Issue{Key: "PROJ-1", ID: "10001"})
		})
		defer srv.Close()

		issue, err := client.CreateIssue(&CreateIssueInput{
			ProjectKey: "PROJ",
			Summary:    "Test issue",
			IssueType:  "Bug",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.Key != "PROJ-1" {
			t.Errorf("expected key PROJ-1, got %s", issue.Key)
		}
	})

	t.Run("all fields", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)

			if fields["description"] == nil {
				t.Error("expected description to be set")
			}
			if fields["priority"] == nil {
				t.Error("expected priority to be set")
			}
			if fields["assignee"] == nil {
				t.Error("expected assignee to be set")
			}
			labels := fields["labels"].([]any)
			if len(labels) != 2 {
				t.Errorf("expected 2 labels, got %d", len(labels))
			}
			comps := fields["components"].([]any)
			if len(comps) != 1 {
				t.Errorf("expected 1 component, got %d", len(comps))
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Issue{Key: "PROJ-2", ID: "10002"})
		})
		defer srv.Close()

		issue, err := client.CreateIssue(&CreateIssueInput{
			ProjectKey:  "PROJ",
			Summary:     "Full issue",
			Description: "A detailed description",
			IssueType:   "Story",
			Priority:    "High",
			Assignee:    "abc123",
			Labels:      []string{"backend", "urgent"},
			Components:  []string{"api"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.Key != "PROJ-2" {
			t.Errorf("expected key PROJ-2, got %s", issue.Key)
		}
	})

	t.Run("API error", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				ErrorMessages: []string{"Project does not exist"},
			})
		})
		defer srv.Close()

		_, err := client.CreateIssue(&CreateIssueInput{
			ProjectKey: "NOPE",
			Summary:    "Fail",
			IssueType:  "Task",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); !contains(got, "Project does not exist") {
			t.Errorf("expected error to mention 'Project does not exist', got: %s", got)
		}
	})
}

func TestGetIssue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/issue/PROJ-42" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			json.NewEncoder(w).Encode(Issue{
				Key: "PROJ-42",
				ID:  "10042",
				Fields: Fields{
					Summary: "Test summary",
					Status:  &Status{Name: "Open"},
					Priority: &Priority{Name: "High"},
					Assignee: &User{DisplayName: "Alice", AccountID: "a1"},
				},
			})
		})
		defer srv.Close()

		issue, err := client.GetIssue("PROJ-42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.Fields.Summary != "Test summary" {
			t.Errorf("expected 'Test summary', got %q", issue.Fields.Summary)
		}
		if issue.Fields.Status.Name != "Open" {
			t.Errorf("expected status 'Open', got %q", issue.Fields.Status.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				ErrorMessages: []string{"Issue does not exist"},
			})
		})
		defer srv.Close()

		_, err := client.GetIssue("NOPE-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); !contains(got, "404") {
			t.Errorf("expected 404 in error, got: %s", got)
		}
	})
}

func TestUpdateIssue(t *testing.T) {
	t.Run("update summary", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/issue/PROJ-10" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)

			if fields["summary"] != "New title" {
				t.Errorf("expected summary 'New title', got %v", fields["summary"])
			}

			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		summary := "New title"
		err := client.UpdateIssue("PROJ-10", &UpdateIssueInput{Summary: &summary})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no fields", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("should not make HTTP request when no fields are set")
		})
		defer srv.Close()

		err := client.UpdateIssue("PROJ-10", &UpdateIssueInput{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); !contains(got, "no fields to update") {
			t.Errorf("expected 'no fields to update', got: %s", got)
		}
	})

	t.Run("multiple fields", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)

			if fields["summary"] == nil {
				t.Error("expected summary")
			}
			if fields["priority"] == nil {
				t.Error("expected priority")
			}
			if fields["labels"] == nil {
				t.Error("expected labels")
			}

			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		summary := "Updated"
		priority := "Low"
		err := client.UpdateIssue("PROJ-10", &UpdateIssueInput{
			Summary:  &summary,
			Priority: &priority,
			Labels:   []string{"refactor"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSearchIssues(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/search/jql" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["jql"] != "project = PROJ" {
				t.Errorf("unexpected JQL: %v", body["jql"])
			}
			if body["maxResults"].(float64) != 10 {
				t.Errorf("unexpected maxResults: %v", body["maxResults"])
			}

			json.NewEncoder(w).Encode(SearchResult{
				Issues: []Issue{
					{Key: "PROJ-1", Fields: Fields{Summary: "First"}},
					{Key: "PROJ-2", Fields: Fields{Summary: "Second"}},
				},
			})
		})
		defer srv.Close()

		result, err := client.SearchIssues("project = PROJ", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Issues) != 2 {
			t.Errorf("expected 2 issues, got %d", len(result.Issues))
		}
	})

	t.Run("empty results", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(SearchResult{
				Issues: []Issue{},
			})
		})
		defer srv.Close()

		result, err := client.SearchIssues("project = EMPTY", 25)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(result.Issues))
		}
	})
}

func TestGetTransitions(t *testing.T) {
	client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/PROJ-5/transitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(TransitionsResponse{
			Transitions: []Transition{
				{ID: "21", Name: "Start Progress", To: Status{Name: "In Progress"}},
				{ID: "31", Name: "Done", To: Status{Name: "Done"}},
			},
		})
	})
	defer srv.Close()

	transitions, err := client.GetTransitions("PROJ-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transitions))
	}
	if transitions[0].To.Name != "In Progress" {
		t.Errorf("expected 'In Progress', got %q", transitions[0].To.Name)
	}
}

func TestTransitionIssue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			transition := body["transition"].(map[string]any)
			if transition["id"] != "21" {
				t.Errorf("expected transition id '21', got %v", transition["id"])
			}

			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		err := client.TransitionIssue("PROJ-5", "21")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddComment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/issue/PROJ-7/comment" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			adf := body["body"].(map[string]any)
			if adf["type"] != "doc" {
				t.Errorf("expected ADF doc type, got %v", adf["type"])
			}

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"10001"}`))
		})
		defer srv.Close()

		comment, err := client.AddComment("PROJ-7", "This is a comment")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if comment.ID != "10001" {
			t.Errorf("expected comment ID 10001, got %s", comment.ID)
		}
	})
}

func TestGetComments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/issue/PROJ-9/comment" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			json.NewEncoder(w).Encode(CommentsResponse{
				Comments: []Comment{
					{ID: "100", Author: &User{DisplayName: "Alice"}, Created: "2025-01-01T00:00:00.000+0000"},
					{ID: "101", Author: &User{DisplayName: "Bob"}, Created: "2025-01-02T00:00:00.000+0000"},
				},
			})
		})
		defer srv.Close()

		comments, err := client.GetComments("PROJ-9")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 2 {
			t.Errorf("expected 2 comments, got %d", len(comments))
		}
		if comments[0].Author.DisplayName != "Alice" {
			t.Errorf("expected author Alice, got %s", comments[0].Author.DisplayName)
		}
	})

	t.Run("empty", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(CommentsResponse{Comments: []Comment{}})
		})
		defer srv.Close()

		comments, err := client.GetComments("PROJ-9")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 0 {
			t.Errorf("expected 0 comments, got %d", len(comments))
		}
	})
}

func TestFindUser(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/api/3/user/search" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("query") != "alice@example.com" {
				t.Errorf("unexpected query: %s", r.URL.Query().Get("query"))
			}

			json.NewEncoder(w).Encode([]User{
				{AccountID: "abc123", DisplayName: "Alice"},
			})
		})
		defer srv.Close()

		accountID, err := client.FindUser("alice@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if accountID != "abc123" {
			t.Errorf("expected abc123, got %s", accountID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]User{})
		})
		defer srv.Close()

		_, err := client.FindUser("nobody@example.com")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !contains(err.Error(), "no user found") {
			t.Errorf("expected 'no user found' error, got: %s", err.Error())
		}
	})
}

func TestResolveAssignee(t *testing.T) {
	t.Run("email triggers lookup", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]User{
				{AccountID: "xyz789", DisplayName: "Bob"},
			})
		})
		defer srv.Close()

		id, err := client.ResolveAssignee("bob@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "xyz789" {
			t.Errorf("expected xyz789, got %s", id)
		}
	})

	t.Run("account ID passed through", func(t *testing.T) {
		client, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("should not make HTTP request for raw account ID")
		})
		defer srv.Close()

		id, err := client.ResolveAssignee("abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "abc123" {
			t.Errorf("expected abc123, got %s", id)
		}
	})
}

func assertBasicAuth(t *testing.T, r *http.Request, expectedUser, expectedPass string) {
	t.Helper()
	user, pass, ok := r.BasicAuth()
	if !ok {
		t.Error("expected basic auth")
		return
	}
	if user != expectedUser {
		t.Errorf("expected user %q, got %q", expectedUser, user)
	}
	if pass != expectedPass {
		t.Errorf("expected password %q, got %q", expectedPass, pass)
	}
}

func assertContentType(t *testing.T, r *http.Request) {
	t.Helper()
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
