package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// healthcheckTimeout bounds the probe so a wedged container fails fast
// instead of hanging the orchestrator's health check.
const healthcheckTimeout = 2 * time.Second

// runHealthcheck probes the local health endpoint and returns a process exit
// code: 0 = healthy, 1 = unhealthy.
//
// It exists so a container image can declare:
//
//	HEALTHCHECK CMD ["/flaggo-web", "healthcheck"]
//
// which works even in a scratch/distroless base image that has no shell,
// curl, or wget. The base URL defaults to http://127.0.0.1:$PORT/healthz and
// can be overridden with HEALTHCHECK_URL.
func runHealthcheck() int {
	addr := os.Getenv("HEALTHCHECK_URL")
	if addr == "" {
		addr = "http://127.0.0.1:" + envOr("PORT", "8080") + "/healthz"
	}

	client := &http.Client{Timeout: healthcheckTimeout}
	resp, err := client.Get(addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck failed: unexpected status %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
