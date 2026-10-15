package riskstore

import "github.com/finguard/platform/pkg/config"

// Select returns a Client based on the runtime configuration.
//
//   - If cfg.BigQueryEmulatorHost is set, returns a *BigQueryClient.
//   - Otherwise returns the in-process MockClient.
//
// The returned Client is safe to share across requests; both
// implementations are stateless.
func Select(cfg *config.Config) Client {
	if cfg != nil && cfg.BigQueryEmulatorHost != "" {
		return &BigQueryClient{
			Host:    cfg.BigQueryEmulatorHost,
			Project: cfg.RiskstoreBQProject,
			Dataset: cfg.RiskstoreBQDataset,
		}
	}
	return MockClient{}
}

// IsMock reports whether the given Client is the in-process mock.
// Used by the HTTP layer to set the per-request Simulated flag without
// adding a method to the public Client interface.
func IsMock(c Client) bool {
	_, ok := c.(MockClient)
	return ok
}
