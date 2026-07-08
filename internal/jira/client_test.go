package jira

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slashpai/jctl/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "https://test.atlassian.net",
		Email:   "user@example.com",
		Token:   "secret",
	}

	client := NewClient(cfg)

	if client.baseURL != cfg.BaseURL {
		t.Errorf("expected baseURL %q, got %q", cfg.BaseURL, client.baseURL)
	}
	if client.email != cfg.Email {
		t.Errorf("expected email %q, got %q", cfg.Email, client.email)
	}
	if client.token != cfg.Token {
		t.Errorf("expected token %q, got %q", cfg.Token, client.token)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
}

func TestDo_SetsAuthAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth")
		}
		if user != "user@test.com" || pass != "token123" {
			t.Errorf("unexpected credentials: %s / %s", user, pass)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected Accept: %s", r.Header.Get("Accept"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient(&config.Config{
		BaseURL: srv.URL,
		Email:   "user@test.com",
		Token:   "token123",
	})

	body, status, err := client.do("GET", "test/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}

	var result map[string]any
	json.Unmarshal(body, &result)
	if result["ok"] != true {
		t.Errorf("unexpected response body: %s", string(body))
	}
}

func TestDo_ConstructsURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/rest/api/3/issue/PROJ-1"
		if r.URL.Path != expected {
			t.Errorf("expected path %q, got %q", expected, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(&config.Config{BaseURL: srv.URL, Email: "e", Token: "t"})
	_, _, err := client.do("GET", "issue/PROJ-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_SendsJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["key"] != "value" {
			t.Errorf("expected body key=value, got %v", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(&config.Config{BaseURL: srv.URL, Email: "e", Token: "t"})
	_, _, err := client.do("POST", "test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_NilBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.ContentLength > 0 {
			t.Error("expected no body for nil input")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(&config.Config{BaseURL: srv.URL, Email: "e", Token: "t"})
	_, _, err := client.do("GET", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
