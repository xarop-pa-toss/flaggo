// Package deploy holds policy tests for the deployment artifacts.
//
// These tests encode the two acceptance criteria for the Flaggo stack:
//
//  1. The stack is exactly four apps: frontend, api, client, db.
//  2. ONLY the frontend publishes a host port. The api, client and db are
//     reachable only on the internal network.
//
// A policy that lives only in a README is a policy that erodes. These run
// in CI next to the unit tests, so a stray `ports:` key fails the build.
package deploy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type composeService struct {
	Image     string   `yaml:"image"`
	Build     *build   `yaml:"build"`
	Ports     []string `yaml:"ports"`
	Expose    []string `yaml:"expose"`
	Networks  []string `yaml:"networks"`
	DependsOn any      `yaml:"depends_on"`
	Volumes   []string `yaml:"volumes"`
	Health    any      `yaml:"healthcheck"`
}

type build struct {
	Context string `yaml:"context"`
	Target  string `yaml:"target"`
}

type composeFile struct {
	Name     string                    `yaml:"name"`
	Services map[string]composeService `yaml:"services"`
	Volumes  map[string]any            `yaml:"volumes"`
	Networks map[string]any            `yaml:"networks"`
}

// repoRoot returns the repository root, derived from this file's location.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine source file location")
	}
	// <root>/internal/deploy/policy_test.go -> <root>
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadCompose(t *testing.T) composeFile {
	t.Helper()
	path := filepath.Join(repoRoot(t), "compose.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read compose.yaml: %v", err)
	}
	var c composeFile
	if err := yaml.Unmarshal(data, &c); err != nil {
		t.Fatalf("parse compose.yaml: %v", err)
	}
	return c
}

func TestStackHasExactlyFourApps(t *testing.T) {
	c := loadCompose(t)

	want := []string{"frontend", "api", "client", "db"}
	if len(c.Services) != len(want) {
		t.Fatalf("stack has %d services (%v), want %d (%v)",
			len(c.Services), keys(c.Services), len(want), want)
	}
	for _, name := range want {
		if _, ok := c.Services[name]; !ok {
			t.Errorf("missing service %q", name)
		}
	}
}

func TestOnlyFrontendPublishesHostPorts(t *testing.T) {
	c := loadCompose(t)

	for name, svc := range c.Services {
		if name == "frontend" {
			if len(svc.Ports) == 0 {
				t.Error("frontend must publish a host port (it is the only entry point)")
			}
			continue
		}
		if len(svc.Ports) != 0 {
			t.Errorf("service %q publishes host ports %v — only the frontend may bind to the host",
				name, svc.Ports)
		}
	}
}

func TestDatabaseIsNotExposed(t *testing.T) {
	c := loadCompose(t)
	db, ok := c.Services["db"]
	if !ok {
		t.Fatal("no db service")
	}
	if len(db.Ports) != 0 {
		t.Errorf("db publishes host ports %v", db.Ports)
	}
	if len(db.Expose) != 0 {
		t.Errorf("db exposes %v — the database needs no inbound port at all", db.Expose)
	}
	if !strings.Contains(db.Image, "postgres") {
		t.Errorf("db image = %q, want a postgres image", db.Image)
	}
}

func TestInternalServicesUseExposeNotPorts(t *testing.T) {
	c := loadCompose(t)
	for _, name := range []string{"api", "client"} {
		svc, ok := c.Services[name]
		if !ok {
			t.Fatalf("missing service %q", name)
		}
		if len(svc.Ports) != 0 {
			t.Errorf("%q must not publish host ports", name)
		}
	}
}

func TestAllServicesShareOnePrivateNetwork(t *testing.T) {
	c := loadCompose(t)
	if len(c.Networks) != 1 {
		t.Fatalf("expected exactly one network, got %d (%v)", len(c.Networks), keys(c.Networks))
	}
	netName := keys(c.Networks)[0]
	for name, svc := range c.Services {
		if len(svc.Networks) == 0 {
			t.Errorf("service %q is not attached to the %q network", name, netName)
			continue
		}
		for _, n := range svc.Networks {
			if n != netName {
				t.Errorf("service %q attached to unexpected network %q", name, n)
			}
		}
	}
}

func TestBuildTargetsMatchServices(t *testing.T) {
	c := loadCompose(t)
	for _, tc := range []struct{ service, target string }{
		{"frontend", "frontend"},
		{"api", "api"},
		{"client", "client"},
	} {
		svc, ok := c.Services[tc.service]
		if !ok {
			t.Fatalf("missing service %q", tc.service)
		}
		if svc.Build == nil {
			t.Errorf("service %q has no build section", tc.service)
			continue
		}
		if svc.Build.Target != tc.target {
			t.Errorf("service %q build target = %q, want %q", tc.service, svc.Build.Target, tc.target)
		}
	}
	// The database is a third-party image, not built from this repo.
	if c.Services["db"].Build != nil {
		t.Error("db must use a published image, not a local build")
	}
}

func TestDatabaseDataIsOnAVolume(t *testing.T) {
	c := loadCompose(t)
	db := c.Services["db"]
	if len(db.Volumes) == 0 {
		t.Fatal("db has no volume — data would be lost on every `docker compose down`")
	}
	if len(c.Volumes) == 0 {
		t.Fatal("no named volumes declared")
	}
}

func TestDependsOnReferencesRealServices(t *testing.T) {
	c := loadCompose(t)
	for name, svc := range c.Services {
		deps, ok := svc.DependsOn.(map[string]any)
		if !ok {
			continue
		}
		for dep := range deps {
			if _, exists := c.Services[dep]; !exists {
				t.Errorf("service %q depends on unknown service %q", name, dep)
			}
		}
	}
}

func TestDockerfileDefinesAllTargets(t *testing.T) {
	path := filepath.Join(repoRoot(t), "Dockerfile")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	dockerfile := string(data)

	for _, target := range []string{"AS api", "AS frontend", "AS client"} {
		if !strings.Contains(dockerfile, target) {
			t.Errorf("Dockerfile has no %q stage", target)
		}
	}
	if !strings.Contains(dockerfile, "CGO_ENABLED=0") {
		t.Error("Dockerfile must build with CGO_ENABLED=0 so the binary runs on distroless/scratch")
	}
	if n := strings.Count(dockerfile, "USER nonroot"); n < 3 {
		t.Errorf("expected all 3 final stages to run as nonroot, found %d", n)
	}
	// The api and frontend are depended on with `condition: service_healthy`,
	// so their images must actually define a HEALTHCHECK.
	if n := strings.Count(dockerfile, "HEALTHCHECK"); n < 2 {
		t.Errorf("expected HEALTHCHECK in the api and frontend stages, found %d", n)
	}
	// The client is a worker: it must not pretend to serve anything.
	clientStage := stageBody(dockerfile, "AS client")
	if strings.Contains(clientStage, "HEALTHCHECK") || strings.Contains(clientStage, "EXPOSE") {
		t.Error("the client stage must not declare HEALTHCHECK or EXPOSE — it serves nothing")
	}
}

// stageBody returns the text of a Dockerfile stage, from its FROM line to the
// next FROM line.
func stageBody(dockerfile, marker string) string {
	parts := strings.Split(dockerfile, "\nFROM ")
	for _, p := range parts {
		if strings.Contains(p, marker) {
			return p
		}
	}
	return ""
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
