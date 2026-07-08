package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateCmd_Run(t *testing.T) {
	t.Run("creates issue with minimal fields", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/issue" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)

			if fields["summary"] != "Test issue" {
				t.Errorf("unexpected summary: %v", fields["summary"])
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-1", "id": "10001"})
		})

		cmd := &CreateCmd{
			Project: "PROJ",
			Summary: "Test issue",
			Type:    "Task",
		}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("resolves email assignee", func(t *testing.T) {
		var gotAssignee string
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/user/search":
				json.NewEncoder(w).Encode([]map[string]string{
					{"accountId": "resolved-id", "displayName": "Alice"},
				})
			case "/rest/api/3/issue":
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				fields := body["fields"].(map[string]any)
				assignee := fields["assignee"].(map[string]any)
				gotAssignee = assignee["accountId"].(string)

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-2", "id": "10002"})
			}
		})

		cmd := &CreateCmd{
			Project:  "PROJ",
			Summary:  "With assignee",
			Type:     "Task",
			Assignee: "alice@example.com",
		}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotAssignee != "resolved-id" {
			t.Errorf("expected resolved-id, got %s", gotAssignee)
		}
	})
}
