package httpx

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

type fakeStore struct {
	flags map[string]model.Flag
}

func newFakeStore() *fakeStore { return &fakeStore{flags: map[string]model.Flag{}} }

func (f *fakeStore) Migrate(ctx context.Context) error { return nil }
func (f *fakeStore) Close() error                      { return nil }
func (f *fakeStore) Create(ctx context.Context, fl model.Flag) error {
	if _, ok := f.flags[fl.Name]; ok {
		return model.ErrAlreadyExists
	}
	f.flags[fl.Name] = fl
	return nil
}
func (f *fakeStore) Get(ctx context.Context, name string) (model.Flag, error) {
	fl, ok := f.flags[name]
	if !ok {
		return model.Flag{}, model.ErrNotFound
	}
	return fl, nil
}
func (f *fakeStore) List(ctx context.Context) ([]model.Flag, error) {
	out := make([]model.Flag, 0, len(f.flags))
	for _, fl := range f.flags {
		out = append(out, fl)
	}
	return out, nil
}
func (f *fakeStore) Update(ctx context.Context, fl model.Flag) error {
	if _, ok := f.flags[fl.Name]; !ok {
		return model.ErrNotFound
	}
	f.flags[fl.Name] = fl
	return nil
}
func (f *fakeStore) Delete(ctx context.Context, name string) error {
	if _, ok := f.flags[name]; !ok {
		return model.ErrNotFound
	}
	delete(f.flags, name)
	return nil
}

func TestFlagsCRUD(t *testing.T) {
	fs := newFakeStore()
	h := NewFlagsHandler(fs, nilLogger())
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"name":"new-ui","enabled":true,"rollout_percent":25}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", rr.Code, rr.Body.String())
	}
	var created model.Flag
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created.Name != "new-ui" || created.RolloutPercent != 25 {
		t.Fatalf("bad created: %+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/flags/new-ui", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("get status %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status %d", rr.Code)
	}

	body = `{"rollout_percent":50}`
	req = httptest.NewRequest(http.MethodPut, "/flags/new-ui/rollout", strings.NewReader(body))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("rollout status %d: %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/flags/new-ui/kill-switch", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("kill status %d", rr.Code)
	}
	var killed model.Flag
	json.Unmarshal(rr.Body.Bytes(), &killed)
	if !killed.KillSwitch {
		t.Fatalf("kill switch not set")
	}

	req = httptest.NewRequest(http.MethodDelete, "/flags/new-ui", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete status %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/flags/new-ui", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
	}
}

func TestFlagsValidation(t *testing.T) {
	fs := newFakeStore()
	h := NewFlagsHandler(fs, nilLogger())
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"name":"x","enabled":true,"rollout_percent":150}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}

	body = `{"name":"dup","enabled":true,"rollout_percent":10}`
	req = httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status %d", rr.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 got %d", rr.Code)
	}
}

func nilLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }
