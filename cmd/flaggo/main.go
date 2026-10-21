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
	httpx "github.com/xarop-pa-toss/flaggo/internal/http"
	"github.com/xarop-pa-toss/flaggo/internal/observability"
	"github.com/xarop-pa-toss/flaggo/internal/store"
)

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
			fmt.Fprintln(os.Stderr, "usage: flaggo [serve|healthcheck|version]")
			os.Exit(2)
		}
	}
	runServer()
}

// version is overridable at build time:
//   go build -ldflags "-X main.version=$(git describe --tags --always)"
var version = "dev"

func runServer() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := newLogger(cfg.LogLevel)
	slog.SetDefault(log)

	metrics := observability.NewMetrics()
	health := httpx.NewHealthHandler()

	// Storage is optional: if DATABASE_URL is unset, the control plane
	// runs in degraded mode (healthz ok, readyz not ready, no flag routes).
	var flagsHandler *httpx.FlagsHandler
	var st store.Store
	if cfg.DatabaseURL != "" {
		s, err := store.NewPostgres(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Error("database connection failed", "err", err)
			health.SetReady(false)
		} else {
			st = s
			if err := st.Migrate(context.Background()); err != nil {
				log.Error("database migration failed", "err", err)
				health.SetReady(false)
				st.Close()
				st = nil
			} else {
				health.SetReady(true)
				flagsHandler = httpx.NewFlagsHandler(st, log)
				log.Info("database ready")
			}
		}
	} else {
		log.Warn("DATABASE_URL not set; running without persistence")
		health.SetReady(false)
	}

	router := httpx.NewRouter(httpx.Deps{
		Health:  health,
		Metrics: metrics,
		Logger:  log,
		Flags:   flagsHandler,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info("shutdown signal received", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	health.SetReady(false)
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
	if st != nil {
		st.Close()
	}
	log.Info("server stopped")
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
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}

// envOr returns the environment variable value or def when unset/empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
