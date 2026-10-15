package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("GCP_PROJECT_ID", "test-project")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.GCPProjectID != "test-project" {
		t.Errorf("GCPProjectID = %q, want %q", cfg.GCPProjectID, "test-project")
	}
	if !cfg.MockMode {
		t.Errorf("MockMode = false, want true (default)")
	}
}

func TestLoadMissingProject(t *testing.T) {
	os.Clearenv()
	if _, err := Load(); err == nil {
		t.Fatal("expected error when GCP_PROJECT_ID is empty")
	}
}

func TestLoadFeatureFlags(t *testing.T) {
	os.Clearenv()
	os.Setenv("GCP_PROJECT_ID", "p")
	os.Setenv("FEATURE_FLAGS", "new_scorecard=on,experimental_llm=true")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.IsFlag("new_scorecard") {
		t.Error("expected new_scorecard=true")
	}
	if !cfg.IsFlag("experimental_llm") {
		t.Error("expected experimental_llm=true")
	}
	if cfg.IsFlag("missing") {
		t.Error("missing flag should be false")
	}
}
