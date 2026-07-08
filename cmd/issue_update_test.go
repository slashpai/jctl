package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUpdateCmd_Run(t *testing.T) {
	t.Run("update summary", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut || r.URL.Path != "/rest/api/3/issue/PROJ-10" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		})

		summary := "New title"
		cmd := &UpdateCmd{IssueKey: "PROJ-10", Summary: &summary}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("add comment", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rest/api/3/issue/PROJ-10/comment" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "99"})
		})

		cmd := &UpdateCmd{IssueKey: "PROJ-10", Comment: "A comment"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no fields or comment", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("should not make HTTP request")
		})

		cmd := &UpdateCmd{IssueKey: "PROJ-10"}
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
