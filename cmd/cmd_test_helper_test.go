package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slashpai/jctl/internal/config"
	"github.com/slashpai/jctl/internal/jira"
)

// withTestClient swaps jiraClient for the duration of the test,
// pointing it at a mock HTTP server with the given handler.
func withTestClient(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := &config.Config{
		BaseURL: srv.URL,
		Email:   "test@example.com",
		Token:   "test-token",
	}

	orig := jiraClient
	jiraClient = func() (*jira.Client, *config.Config, error) {
		return jira.NewClient(cfg), cfg, nil
	}
	t.Cleanup(func() { jiraClient = orig })
}
