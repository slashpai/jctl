package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		BaseURL: "https://test.atlassian.net",
		Email:   "user@example.com",
		Token:   "secret-token",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path := filepath.Join(tmpDir, ".config", "jctl", "config.yaml")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected file permissions 0600, got %o", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", loaded.BaseURL, cfg.BaseURL)
	}
	if loaded.Email != cfg.Email {
		t.Errorf("Email = %q, want %q", loaded.Email, cfg.Email)
	}
	if loaded.Token != cfg.Token {
		t.Errorf("Token = %q, want %q", loaded.Token, cfg.Token)
	}
}

func TestLoad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when config file missing")
	}
	if got := err.Error(); !containsStr(got, "config not found") {
		t.Errorf("expected 'config not found' error, got: %s", got)
	}
}

func TestLoad_IncompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	dir := filepath.Join(tmpDir, ".config", "jctl")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("base_url: https://test.atlassian.net\n"), 0o600)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for incomplete config")
	}
	if got := err.Error(); !containsStr(got, "incomplete config") {
		t.Errorf("expected 'incomplete config' error, got: %s", got)
	}
}

func TestLoad_RejectsHTTP(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	dir := filepath.Join(tmpDir, ".config", "jctl")
	os.MkdirAll(dir, 0o700)
	data := "base_url: http://insecure.example.com\nemail: a@b.com\ntoken: tok\n"
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(data), 0o600)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for HTTP base_url")
	}
	if got := err.Error(); !containsStr(got, "HTTPS") {
		t.Errorf("expected HTTPS error, got: %s", got)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		BaseURL: "https://original.atlassian.net",
		Email:   "original@example.com",
		Token:   "original-token",
	}
	Save(cfg)

	t.Setenv("JCTL_BASE_URL", "https://override.atlassian.net")
	t.Setenv("JCTL_EMAIL", "override@example.com")
	t.Setenv("JCTL_TOKEN", "override-token")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.BaseURL != "https://override.atlassian.net" {
		t.Errorf("BaseURL not overridden: %s", loaded.BaseURL)
	}
	if loaded.Email != "override@example.com" {
		t.Errorf("Email not overridden: %s", loaded.Email)
	}
	if loaded.Token != "override-token" {
		t.Errorf("Token not overridden: %s", loaded.Token)
	}
}

func TestConfigFilePath(t *testing.T) {
	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error: %v", err)
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(path))
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
