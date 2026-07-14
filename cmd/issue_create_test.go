package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateCmd_Run(t *testing.T) {
	t.Run("creates issue with minimal fields", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/myself":
				json.NewEncoder(w).Encode(map[string]string{
					"accountId":   "current-user-id",
					"displayName": "Alice",
				})
			case "/rest/api/3/issue":
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
				}

				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				fields := body["fields"].(map[string]any)

				if fields["summary"] != "Test issue" {
					t.Errorf("unexpected summary: %v", fields["summary"])
				}
				assignee := fields["assignee"].(map[string]any)
				if assignee["accountId"] != "current-user-id" {
					t.Errorf("expected default assignee current-user-id, got %v", assignee["accountId"])
				}

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-1", "id": "10001"})
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
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

	t.Run("assigns to me", func(t *testing.T) {
		var gotAssignee string
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/myself":
				json.NewEncoder(w).Encode(map[string]string{
					"accountId":   "current-user-id",
					"displayName": "Alice",
				})
			case "/rest/api/3/issue":
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				fields := body["fields"].(map[string]any)
				assignee := fields["assignee"].(map[string]any)
				gotAssignee = assignee["accountId"].(string)

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-3", "id": "10003"})
			}
		})

		cmd := &CreateCmd{
			Project:  "PROJ",
			Summary:  "Assigned to me",
			Type:     "Task",
			Assignee: "me",
		}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotAssignee != "current-user-id" {
			t.Errorf("expected current-user-id, got %s", gotAssignee)
		}
	})

	t.Run("leaves issue unassigned with none", func(t *testing.T) {
		withTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rest/api/3/issue" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fields := body["fields"].(map[string]any)
			if fields["assignee"] != nil {
				t.Errorf("expected no assignee, got %v", fields["assignee"])
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"key": "PROJ-4", "id": "10004"})
		})

		cmd := &CreateCmd{
			Project:  "PROJ",
			Summary:  "Unassigned",
			Type:     "Task",
			Assignee: "none",
		}
		if err := cmd.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
