package httpx

import (
	"log/slog"
	"net/http"

	"github.com/xarop-pa-toss/flaggo/internal/http/middleware"
	"github.com/xarop-pa-toss/flaggo/internal/observability"
)

// Deps holds dependencies for the HTTP layer.
type Deps struct {
	Health  *HealthHandler
	Metrics *observability.Metrics
	Logger  *slog.Logger
	Flags   *FlagsHandler
}

// NewRouter builds the HTTP mux with middleware chain applied.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", d.Health.Healthz)
	mux.HandleFunc("GET /readyz", d.Health.Readyz)
	mux.Handle("GET /metrics", observability.Handler())
	if d.Flags != nil {
		d.Flags.Register(mux)
	}

	var h http.Handler = mux
	h = middleware.Metrics(d.Metrics)(h)
	h = middleware.Logging(d.Logger)(h)
	h = middleware.Recovery(d.Logger)(h)
	h = middleware.RequestID(h)
	return h
}
