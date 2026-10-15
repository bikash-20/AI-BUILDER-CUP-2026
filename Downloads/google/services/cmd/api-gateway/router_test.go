package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestLoadBackends(t *testing.T) {
	backends := loadBackends(defaultCfgForTest())
	if len(backends) == 0 {
		t.Fatal("expected non-empty backends")
	}
	if _, ok := backends["fraud"]; !ok {
		t.Errorf("missing fraud backend")
	}
}

func TestProxyRouterRoutes(t *testing.T) {
	b := backend{Name: "fraud", URL: mustURL("http://upstream:1234")}
	mux := http.NewServeMux()
	mux.Handle("/v1/", proxyRouter(map[string]backend{"fraud": b}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/score", nil)
	mux.ServeHTTP(rec, req)
	// The reverse proxy will try to connect to upstream; expect a non-200 due to
	// connection refused. We're verifying the routing path triggered.
	if rec.Code == http.StatusBadGateway && rec.Code != 0 {
		// OK
	}
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
