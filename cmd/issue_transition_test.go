package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/slashpai/jctl/internal/jira"
)

func TestTransitionCmd_Run(t *testing.T) {
	t.Run("list transitions", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(jira.TransitionsResponse{
				Transitions: []jira.Transition{
					{ID: "21", Name: "Start", To: jira.Status{Name: "In Progress"}},
					{ID: "31", Name: "Done", To: jira.Status{Name: "Done"}},
				},
			})
		})

		cmd := &TransitionCmd{IssueKey: "PROJ-5", List: true}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no flags shows transitions with hint", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(jira.TransitionsResponse{
				Transitions: []jira.Transition{
					{ID: "21", Name: "Start", To: jira.Status{Name: "In Progress"}},
				},
			})
		})

		cmd := &TransitionCmd{IssueKey: "PROJ-5"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("transition by status", func(t *testing.T) {
		callCount := 0
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			callCount++
			switch {
			case r.Method == http.MethodGet:
				json.NewEncoder(w).Encode(jira.TransitionsResponse{
					Transitions: []jira.Transition{
						{ID: "21", Name: "Start", To: jira.Status{Name: "In Progress"}},
					},
				})
			case r.Method == http.MethodPost:
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				transition := body["transition"].(map[string]any)
				if transition["id"] != "21" {
					t.Errorf("expected transition id 21, got %v", transition["id"])
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})

		cmd := &TransitionCmd{IssueKey: "PROJ-5", Status: "In Progress"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("transition not found", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(jira.TransitionsResponse{
				Transitions: []jira.Transition{
					{ID: "21", Name: "Start", To: jira.Status{Name: "In Progress"}},
				},
			})
		})

		cmd := &TransitionCmd{IssueKey: "PROJ-5", Status: "Nonexistent"}
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty transitions", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(jira.TransitionsResponse{Transitions: []jira.Transition{}})
		})

		cmd := &TransitionCmd{IssueKey: "PROJ-5", List: true}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
