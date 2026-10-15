package main

import (
	"os"

	"github.com/finguard/platform/pkg/config"
)

// defaultCfgForTest returns a minimal config for tests, setting project id.
func defaultCfgForTest() *config.Config {
	os.Setenv("GCP_PROJECT_ID", "test-project")
	c, _ := config.Load()
	if c == nil {
		c = &config.Config{Env: "dev"}
	}
	return c
}
