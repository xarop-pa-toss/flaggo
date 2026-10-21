package web_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	httpx "github.com/xarop-pa-toss/flaggo/internal/http"
	"github.com/xarop-pa-toss/flaggo/internal/observability"
	"github.com/xarop-pa-toss/flaggo/internal/store"
	"github.com/xarop-pa-toss/flaggo/internal/web"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newStack starts an in-process Control Plane API plus a frontend/gateway
// wired to it, exactly as the compose stack wires them over the network.
func newStack(t *testing.T) (api *httptest.Server, front *httptest.Server) {
	t.Helper()

	st := store.NewMemory()
	apiRouter := httpx.NewRouter(httpx.Deps{
		Health:  httpx.NewHealthHandler(),
		Metrics: observability.NewMetrics(prometheus.NewRegistry()),
		Logger:  testLogger(),
		Flags:   httpx.NewFlagsHandler(st, testLogger()),
	})
	api = httptest.NewServer(apiRouter)
	t.Cleanup(api.Close)

	frontHandler, err := web.NewHandler(web.Deps{
		BackendURL: api.URL,
		Logger:     testLogger(),
		Metrics:    observability.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatalf("web.NewHandler: %v", err)
	}
	front = httptest.NewServer(frontHandler)
	t.Cleanup(front.Close)
	return api, front
}

func TestDashboardShellIsServed(t *testing.T) {
	_, front := newStack(t)

	resp, err := http.Get(front.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("GET / content-type = %q, want text/html", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Flaggo") {
		t.Fatal("dashboard HTML does not mention Flaggo")
	}
}

func TestStaticAssetsAreServed(t *testing.T) {
	_, front := newStack(t)

	for _, asset := range []string{"/static/app.js", "/static/style.css"} {
		resp, err := http.Get(front.URL + asset)
		if err != nil {
			t.Fatalf("GET %s: %v", asset, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", asset, resp.StatusCode)
		}
		if len(body) == 0 {
			t.Fatalf("GET %s returned empty body", asset)
		}
	}
}

func TestGatewayProxiesFlagAPI(t *testing.T) {
	_, front := newStack(t)

	// create through the gateway
	body := `{"name":"checkout-v2","description":"new checkout","enabled":true,"rollout_percent":40}`
	resp, err := http.Post(front.URL+"/api/flags", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/flags: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /api/flags status = %d (%s), want 201", resp.StatusCode, payload)
	}

	// list through the gateway
	resp, err = http.Get(front.URL + "/api/flags")
	if err != nil {
		t.Fatalf("GET /api/flags: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/flags status = %d, want 200", resp.StatusCode)
	}
	var flags []struct {
		Name           string `json:"name"`
		Enabled        bool   `json:"enabled"`
		RolloutPercent int    `json:"rollout_percent"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&flags); err != nil {
		t.Fatalf("decode flags: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(flags))
	}
	if flags[0].Name != "checkout-v2" || !flags[0].Enabled || flags[0].RolloutPercent != 40 {
		t.Fatalf("unexpected flag: %+v", flags[0])
	}
}

func TestGatewayProxiesSubresourcePaths(t *testing.T) {
	_, front := newStack(t)

	body := `{"name":"killable","enabled":true,"rollout_percent":100}`
	resp, err := http.Post(front.URL+"/api/flags", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	resp.Body.Close()

	// /api/flags/<name>/rollout must reach the backend's /flags/<name>/rollout
	req, _ := http.NewRequest(http.MethodPut, front.URL+"/api/flags/killable/rollout",
		strings.NewReader(`{"rollout_percent":5}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT rollout: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT rollout status = %d (%s), want 200", resp.StatusCode, payload)
	}

	// and it really changed on the backend
	resp, err = http.Get(front.URL + "/api/flags/killable")
	if err != nil {
		t.Fatalf("GET flag: %v", err)
	}
	defer resp.Body.Close()
	var f struct {
		RolloutPercent int `json:"rollout_percent"`
	}
	json.NewDecoder(resp.Body).Decode(&f)
	if f.RolloutPercent != 5 {
		t.Fatalf("rollout = %d, want 5", f.RolloutPercent)
	}
}

func TestGatewayPropagatesRequestID(t *testing.T) {
	_, front := newStack(t)

	req, _ := http.NewRequest(http.MethodGet, front.URL+"/api/flags", nil)
	req.Header.Set("X-Request-ID", "trace-me-123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-Request-ID"); got != "trace-me-123" {
		t.Fatalf("X-Request-ID = %q, want %q (id must survive the proxy hop)", got, "trace-me-123")
	}
}

func TestHealthAndReadiness(t *testing.T) {
	_, front := newStack(t)

	for path, want := range map[string]int{"/healthz": 200, "/readyz": 200} {
		resp, err := http.Get(front.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Fatalf("GET %s status = %d, want %d", path, resp.StatusCode, want)
		}
	}
}

func TestReadyzFailsWhenBackendIsDown(t *testing.T) {
	// Frontend pointed at a dead backend: liveness still ok, readiness not.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	backendURL := dead.URL
	dead.Close() // now nothing is listening

	handler, err := web.NewHandler(web.Deps{
		BackendURL: backendURL,
		Logger:     testLogger(),
		Metrics:    observability.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatalf("web.NewHandler: %v", err)
	}
	front := httptest.NewServer(handler)
	defer front.Close()

	resp, err := http.Get(front.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz status = %d, want 200 (liveness must not depend on the backend)", resp.StatusCode)
	}

	resp, err = http.Get(front.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/readyz status = %d, want 503", resp.StatusCode)
	}
}

func TestBackendUnreachableReturns502(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	backendURL := dead.URL
	dead.Close()

	handler, err := web.NewHandler(web.Deps{
		BackendURL: backendURL,
		Logger:     testLogger(),
		Metrics:    observability.NewMetrics(prometheus.NewRegistry()),
	})
	if err != nil {
		t.Fatalf("web.NewHandler: %v", err)
	}
	front := httptest.NewServer(handler)
	defer front.Close()

	resp, err := http.Get(front.URL + "/api/flags")
	if err != nil {
		t.Fatalf("GET /api/flags: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
}
