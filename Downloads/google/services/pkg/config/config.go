// Package config loads FinGuard service configuration from environment variables.
//
// All FinGuard services follow the same 12-factor configuration pattern.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration for a FinGuard service.
type Config struct {
	ServiceName string
	Env         string // dev | staging | prod
	Port        int

	GCPProjectID    string
	GCPRegion       string
	GCPEndpointBase string // optional override for emulator

	MockMode bool

	VertexEndpointID string

	BigQueryProject   string
	BigQueryDataset   string
	BigQueryFeatureMV string

	PubSubTopicScored string
	PubSubTopicAlerts string
	PubSubSubAlerts   string

	AlloyDBDSN string

	// BigQueryEmulatorHost, if set, enables the live BigQuery path for
	// riskstore. Format: "http://127.0.0.1:9050". When empty, riskstore
	// uses its in-process MockClient and the response carries
	// Simulated=true.
	BigQueryEmulatorHost string
	// RiskstoreBQProject / RiskstoreBQDataset select the project + dataset
	// the riskstore BigQuery client queries. Defaults match the goccy
	// emulator defaults wired in scripts/restart_all.sh.
	RiskstoreBQProject string
	RiskstoreBQDataset string

	DocAIBucket string
	DLQBucket   string

	OTelExporter string

	MockVertex       bool
	FeatureFlags     map[string]bool

	// TimeoutD is the score-timeout duration (named to avoid colliding with the
	// accessor method).
	TimeoutD time.Duration
	BigQueryTimeoutD time.Duration
	VertexTimeoutD   time.Duration
	VertexFallback   string
}

// ScoreTimeout returns the configured score-timeout duration.
func (c *Config) ScoreTimeout() time.Duration { return c.TimeoutD }

// BigQueryTimeout returns the configured BigQuery timeout.
func (c *Config) BigQueryTimeout() time.Duration { return c.BigQueryTimeoutD }

// VertexTimeout returns the configured Vertex AI timeout.
func (c *Config) VertexTimeout() time.Duration { return c.VertexTimeoutD }

// Load reads env vars and returns a populated Config.
func Load() (*Config, error) {
	cfg := &Config{
		ServiceName:        envDefault("SERVICE_NAME", "finguard-service"),
		Env:                envDefault("ENV", "dev"),
		Port:               envInt("PORT", 8080),
		GCPProjectID:       envDefault("GCP_PROJECT_ID", "finguard-dev"),
		GCPRegion:          envDefault("GCP_REGION", "asia-south1"),
		GCPEndpointBase:    os.Getenv("GCP_ENDPOINT_BASE"),
		MockMode:           envBool("MOCK_MODE", true),
		VertexEndpointID:   os.Getenv("VERTEX_ENDPOINT_ID"),
		BigQueryProject:    envDefault("BIGQUERY_PROJECT", "finguard-dev"),
		BigQueryDataset:    envDefault("BIGQUERY_DATASET", "fin"),
		BigQueryFeatureMV:  envDefault("BIGQUERY_FEATURE_MV", "fin.customer_features_mv"),
		PubSubTopicScored:  envDefault("PUBSUB_TOPIC_SCORED", "txn-scored"),
		PubSubTopicAlerts:  envDefault("PUBSUB_TOPIC_ALERTS", "alerts"),
		PubSubSubAlerts:    envDefault("PUBSUB_SUB_ALERTS", "alerts-sub"),
		AlloyDBDSN:         os.Getenv("ALLOYDB_DSN"),
		BigQueryEmulatorHost: os.Getenv("BIGQUERY_EMULATOR_HOST"),
		RiskstoreBQProject: envDefault("RISKSTORE_BQ_PROJECT", "finguard-dev"),
		RiskstoreBQDataset: envDefault("RISKSTORE_BQ_DATASET", "finguard_dev"),
		DocAIBucket:        envDefault("DOCAI_BUCKET", "finguard-documents"),
		DLQBucket:          envDefault("DLQ_BUCKET", "finguard-dlq"),
		OTelExporter:       envDefault("OTEL_EXPORTER", "none"),
		MockVertex:         envBool("MOCK_VERTEX", true),
		FeatureFlags:       parseFeatureFlags(os.Getenv("FEATURE_FLAGS")),
		TimeoutD:           envDuration("SCORE_TIMEOUT", 450*time.Millisecond),
		BigQueryTimeoutD:   envDuration("BIGQUERY_TIMEOUT", 200*time.Millisecond),
		VertexTimeoutD:     envDuration("VERTEX_TIMEOUT", 250*time.Millisecond),
		VertexFallback:     envDefault("VERTEX_FALLBACK", "rule_based"),
	}

	if cfg.GCPProjectID == "" || os.Getenv("GCP_PROJECT_ID") == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid PORT %d", cfg.Port)
	}
	return cfg, nil
}

// IsFlag reports whether a feature flag is enabled.
func (c *Config) IsFlag(name string) bool { return c.FeatureFlags[name] }

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	}
	return false
}

func parseFeatureFlags(raw string) map[string]bool {
	out := make(map[string]bool)
	if raw == "" {
		return out
	}
	for _, kv := range strings.Split(raw, ",") {
		kv = strings.TrimSpace(kv)
		if kv == "" {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			out[parts[0]] = parseBool(parts[1])
		} else {
			out[parts[0]] = true
		}
	}
	return out
}
