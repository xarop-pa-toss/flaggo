// Command client is the Flaggo Data Plane example.
//
// It polls the Control Plane for the flag set and evaluates every flag in
// memory for one user key. When a flag's effective value changes, it prints a
// CHANGE line — so you can flip a flag in the dashboard and watch this process
// react within one poll interval, without restarting.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/xarop-pa-toss/flaggo/internal/eval"
	"github.com/xarop-pa-toss/flaggo/internal/model"
)

func main() {
	baseURL := envOr("CONTROL_PLANE_URL", "http://localhost:8080")
	interval := durOr("POLL_INTERVAL", 2*time.Second)
	userKey := envOr("USER_KEY", "demo-user")
	only := os.Getenv("FLAG_NAME")

	fmt.Printf("data plane polling %s every %s as user %q\n", baseURL, interval, userKey)
	if only != "" {
		fmt.Printf("filtering to flag %q\n", only)
	}
	fmt.Println(strings.Repeat("-", 72))

	prev := map[string]bool{}
	seen := map[string]bool{}

	for {
		flags, err := fetchFlags(baseURL, only)
		if err != nil {
			fmt.Printf("[poll error] %v\n", err)
			time.Sleep(interval)
			continue
		}
		sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })

		if len(flags) == 0 {
			fmt.Println("[waiting] no flags defined yet — create one in the dashboard")
		}

		for _, f := range flags {
			on := eval.Evaluate(f, userKey)
			state := "OFF"
			if on {
				state = "ON "
			}
			detail := fmt.Sprintf("enabled=%v rollout=%3d%% kill=%v", f.Enabled, f.RolloutPercent, f.KillSwitch)

			if seen[f.Name] && prev[f.Name] != on {
				from, to := "OFF", "ON"
				if !on {
					from, to = "ON", "OFF"
				}
				fmt.Printf("[CHANGE] %-16s %s -> %s  (%s)\n", f.Name, from, to, detail)
			} else {
				fmt.Printf("[state]  %-16s %s  (%s)\n", f.Name, state, detail)
			}
			prev[f.Name] = on
			seen[f.Name] = true
		}

		time.Sleep(interval)
	}
}

func fetchFlags(baseURL, only string) ([]model.Flag, error) {
	path := "/flags"
	if only != "" {
		path = "/flags/" + only
	}
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	if only != "" {
		var f model.Flag
		if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
			return nil, err
		}
		return []model.Flag{f}, nil
	}

	var flags []model.Flag
	if err := json.NewDecoder(resp.Body).Decode(&flags); err != nil {
		return nil, err
	}
	return flags, nil
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
