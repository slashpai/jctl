package cmd

import "testing"

func TestBuildJQL(t *testing.T) {
	t.Run("empty returns empty", func(t *testing.T) {
		got := buildJQL(&ListCmd{})
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("project only", func(t *testing.T) {
		got := buildJQL(&ListCmd{Project: "PROJ"})
		want := `project = "PROJ" ORDER BY updated DESC`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("assignee me", func(t *testing.T) {
		got := buildJQL(&ListCmd{Project: "PROJ", Assignee: "me"})
		if !containsStr(got, "assignee = currentUser()") {
			t.Errorf("expected currentUser() for 'me', got %q", got)
		}
		if !containsStr(got, `project = "PROJ"`) {
			t.Errorf("expected project clause, got %q", got)
		}
	})

	t.Run("assignee specific user", func(t *testing.T) {
		got := buildJQL(&ListCmd{Assignee: "john.doe"})
		if !containsStr(got, `assignee = "john.doe"`) {
			t.Errorf("expected quoted assignee, got %q", got)
		}
	})

	t.Run("status filter", func(t *testing.T) {
		got := buildJQL(&ListCmd{Project: "PROJ", Status: "In Progress"})
		if !containsStr(got, `status = "In Progress"`) {
			t.Errorf("expected status clause, got %q", got)
		}
	})

	t.Run("issue type filter", func(t *testing.T) {
		got := buildJQL(&ListCmd{Project: "PROJ", Type: "Bug"})
		if !containsStr(got, `issuetype = "Bug"`) {
			t.Errorf("expected issuetype clause, got %q", got)
		}
	})

	t.Run("all filters combined", func(t *testing.T) {
		got := buildJQL(&ListCmd{
			Project:  "PROJ",
			Assignee: "me",
			Status:   "Done",
			Type:     "Story",
		})

		if !containsStr(got, `project = "PROJ"`) {
			t.Errorf("missing project, got %q", got)
		}
		if !containsStr(got, "assignee = currentUser()") {
			t.Errorf("missing assignee, got %q", got)
		}
		if !containsStr(got, `status = "Done"`) {
			t.Errorf("missing status, got %q", got)
		}
		if !containsStr(got, `issuetype = "Story"`) {
			t.Errorf("missing issuetype, got %q", got)
		}
		if !containsStr(got, "ORDER BY updated DESC") {
			t.Errorf("missing ORDER BY, got %q", got)
		}
		if !containsStr(got, " AND ") {
			t.Errorf("expected AND between clauses, got %q", got)
		}
	})
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
