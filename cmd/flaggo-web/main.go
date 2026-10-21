// Command flaggo-web is the Flaggo Control Plane Frontend: the management
// dashboard and the API gateway in front of the Control Plane Backend.
//
// It is the only Flaggo component that needs a host-published port.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/xarop-pa-toss/flaggo/internal/config"
	"github.com/xarop-pa-toss/flaggo/internal/observability"
	"github.com/xarop-pa-toss/flaggo/internal/web"
)

// version is overridable at build time:
//
//	go build -ldflags "-X main.version=$(git describe --tags --always)"
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "healthcheck":
			os.Exit(runHealthcheck())
		case "version":
			fmt.Println(version)
			return
		case "serve":
			// default mode, explicit for clarity
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "usage: flaggo-web [serve|healthcheck|version]")
			os.Exit(2)
		}
	}
	runServer()
}

func runServer() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := newLogger(cfg.LogLevel)
	slog.SetDefault(log)

	backendURL := envOr("BACKEND_URL", "http://api:8080")

	metrics := observability.NewDefaultMetrics()
	handler, err := web.NewHandler(web.Deps{
		BackendURL: backendURL,
		Logger:     log,
		Metrics:    metrics,
	})
	if err != nil {
		log.Error("cannot start frontend", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Info("frontend starting", "port", cfg.Port, "backend", backendURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("frontend failed", "err", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info("shutdown signal received", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
	log.Info("frontend stopped")
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

// envOr returns the environment variable value or def when unset/empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
