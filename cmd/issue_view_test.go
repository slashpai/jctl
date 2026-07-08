package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/slashpai/jctl/internal/jira"
)

func TestViewCmd_Run(t *testing.T) {
	t.Run("displays issue details", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rest/api/3/issue/PROJ-42" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			json.NewEncoder(w).Encode(jira.Issue{
				Key: "PROJ-42",
				Fields: jira.Fields{
					Summary:  "Test summary",
					Status:   &jira.Status{Name: "Open"},
					Priority: &jira.Priority{Name: "High"},
				},
			})
		})

		cmd := &ViewCmd{IssueKey: "PROJ-42"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with comments flag", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/PROJ-42":
				json.NewEncoder(w).Encode(jira.Issue{
					Key:    "PROJ-42",
					Fields: jira.Fields{Summary: "Test"},
				})
			case "/rest/api/3/issue/PROJ-42/comment":
				json.NewEncoder(w).Encode(jira.CommentsResponse{
					Comments: []jira.Comment{
						{ID: "1", Author: &jira.User{DisplayName: "Alice"}, Created: "2025-01-01"},
					},
				})
			}
		})

		cmd := &ViewCmd{IssueKey: "PROJ-42", Comments: true}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("issue not found", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"errorMessages": []string{"Issue does not exist"},
			})
		})

		cmd := &ViewCmd{IssueKey: "NOPE-1"}
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
