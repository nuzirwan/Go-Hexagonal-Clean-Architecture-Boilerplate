package e2e

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuzirwan/go-boilerplate/pkg/health"
)

// TestHealthEndpoint verifies the health probe works end-to-end.
func TestHealthEndpoint(t *testing.T) {
	healthHandler := health.NewHandler(map[string]health.Checker{})

	mux := http.NewServeMux()
	healthHandler.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	data, _ := io.ReadAll(resp.Body)
	sonic.Unmarshal(data, &body)
	assert.Equal(t, "ok", body["status"])
}
