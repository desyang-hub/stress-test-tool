package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveEnvVars(t *testing.T) {
	os.Setenv("TEST_HOST", "https://example.com")
	os.Setenv("TEST_TOKEN", "Bearer abc123")
	defer os.Unsetenv("TEST_HOST")
	defer os.Unsetenv("TEST_TOKEN")

	cfg := &TestConfig{
		Name: "test",
		DefaultRequest: DefaultRequest{
			URL:     "${TEST_HOST}/api",
			AuthToken: "$TEST_TOKEN",
		},
		Stages: []Stage{
			{
				Name: "load",
			},
		},
		Requests: []Request{
			{
				ID:  "api-call",
				URL: "${TEST_HOST}/users",
			},
		},
	}

	ResolveEnvVars(cfg)

	if cfg.DefaultRequest.URL != "https://example.com/api" {
		t.Errorf("URL = %q, want https://example.com/api", cfg.DefaultRequest.URL)
	}
	if cfg.DefaultRequest.AuthToken != "Bearer abc123" {
		t.Errorf("AuthToken = %q, want Bearer abc123", cfg.DefaultRequest.AuthToken)
	}
	if cfg.Requests[0].URL != "https://example.com/users" {
		t.Errorf("request URL = %q, want https://example.com/users", cfg.Requests[0].URL)
	}
}

func TestResolveEnvVarsUnresolved(t *testing.T) {
	os.Unsetenv("MISSING_VAR_XYZ")

	cfg := &TestConfig{
		Name: "test",
		DefaultRequest: DefaultRequest{
			URL: "${MISSING_VAR_XYZ}/api",
		},
	}

	ResolveEnvVars(cfg)

	// Should be left as-is when env var is missing
	if cfg.DefaultRequest.URL != "${MISSING_VAR_XYZ}/api" {
		t.Errorf("URL = %q, want ${MISSING_VAR_XYZ}/api (unresolved)", cfg.DefaultRequest.URL)
	}
}

func TestValidateNoStages(t *testing.T) {
	cfg := &TestConfig{Name: "test"}
	errs := Validate(cfg)
	if len(errs) == 0 {
		t.Error("expected error for missing stages")
	}
}

func TestValidateBadMethod(t *testing.T) {
	cfg := &TestConfig{
		Name: "test",
		Stages: []Stage{{Duration: Duration{Duration: time.Second}, Concurrency: 1}},
		Requests: []Request{{
			ID:     "bad",
			Method: "INVALID",
			URL:    "http://example.com",
		}},
	}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if e != nil && contains(e.Error(), "method") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for invalid method")
	}
}

func TestValidateMissingRequestID(t *testing.T) {
	cfg := &TestConfig{
		Name: "test",
		Stages: []Stage{{Duration: Duration{Duration: time.Second}, Concurrency: 1}},
		Requests: []Request{{
			URL: "http://example.com",
		}},
	}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if e != nil && contains(e.Error(), "id is required") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing request id")
	}
}

func TestValidateBadFormat(t *testing.T) {
	cfg := &TestConfig{
		Name: "test",
		Stages: []Stage{{Duration: Duration{Duration: time.Second}, Concurrency: 1}},
		Output: OutputConfig{
			Formats: []string{"console", "invalid_format"},
		},
	}
	errs := Validate(cfg)
	found := false
	for _, e := range errs {
		if e != nil && contains(e.Error(), "format") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for invalid format")
	}
}

func TestLoadValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `
name: Test Load
stages:
  - duration: 10s
    concurrency: 5
    rate_limit: 10
output:
  formats: [console, html]
requests:
  - id: health
    method: GET
    url: http://localhost/health
`
	path := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Name != "Test Load" {
		t.Errorf("Name = %q, want Test Load", cfg.Name)
	}
	if len(cfg.Stages) != 1 {
		t.Errorf("Stages count = %d, want 1", len(cfg.Stages))
	}
	if cfg.Stages[0].Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5", cfg.Stages[0].Concurrency)
	}
	if len(cfg.Requests) != 1 {
		t.Errorf("Requests count = %d, want 1", len(cfg.Requests))
	}
}

func TestValidateValid(t *testing.T) {
	cfg := &TestConfig{
		Name: "valid test",
		Stages: []Stage{{Duration: Duration{Duration: time.Second}, Concurrency: 1}},
		Output: OutputConfig{Formats: []string{"console"}},
	}
	errs := Validate(cfg)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
