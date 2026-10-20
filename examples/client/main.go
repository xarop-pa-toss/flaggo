package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/xarop-pa-toss/flaggo/internal/eval"
	"github.com/xarop-pa-toss/flaggo/internal/model"
)

func main() {
	baseURL := envOr("CONTROL_PLANE_URL", "http://localhost:8080")
	interval := durOr("POLL_INTERVAL", 2*time.Second)
	userKey := envOr("USER_KEY", "demo-user")
	flagName := envOr("FLAG_NAME", "new-ui")

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	log.Info("starting data-plane client", "control_plane", baseURL, "interval", interval, "user", userKey, "flag", flagName)

	for {
		flag, err := fetchFlag(baseURL, flagName)
		if err != nil {
			log.Warn("poll failed", "err", err)
		} else {
			on := eval.Evaluate(flag, userKey)
			fmt.Printf("flag=%s enabled=%v rollout=%d%% kill=%v => for user %q: %v\n",
				flag.Name, flag.Enabled, flag.RolloutPercent, flag.KillSwitch, userKey, on)
		}
		time.Sleep(interval)
	}
}

func fetchFlag(baseURL, name string) (model.Flag, error) {
	resp, err := http.Get(baseURL + "/flags/" + name)
	if err != nil {
		return model.Flag{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.Flag{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var f model.Flag
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return model.Flag{}, err
	}
	return f, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durOr(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
