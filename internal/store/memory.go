package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

// MemoryStore is an in-memory Store implementation.
//
// It exists for two reasons:
//  1. Tests that need a Store without a database.
//  2. Local development of the API and dashboard before a Postgres exists
//     (STORE=memory). Postgres remains the production store.
type MemoryStore struct {
	mu    sync.RWMutex
	flags map[string]model.Flag
}

// NewMemory returns an empty in-memory store.
func NewMemory() *MemoryStore {
	return &MemoryStore{flags: map[string]model.Flag{}}
}

// Migrate is a no-op for the in-memory store.
func (s *MemoryStore) Migrate(ctx context.Context) error { return nil }

// Close is a no-op for the in-memory store.
func (s *MemoryStore) Close() error { return nil }

// Create inserts a flag.
func (s *MemoryStore) Create(ctx context.Context, f model.Flag) error {
	if err := f.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Name]; ok {
		return model.ErrAlreadyExists
	}
	now := time.Now().UTC()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = now
	}
	s.flags[f.Name] = f
	return nil
}

// Get returns a flag by name.
func (s *MemoryStore) Get(ctx context.Context, name string) (model.Flag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[name]
	if !ok {
		return model.Flag{}, model.ErrNotFound
	}
	return f, nil
}

// List returns all flags ordered by name.
func (s *MemoryStore) List(ctx context.Context) ([]model.Flag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Update replaces an existing flag.
func (s *MemoryStore) Update(ctx context.Context, f model.Flag) error {
	if err := f.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Name]; !ok {
		return model.ErrNotFound
	}
	existing, _ := s.get(f.Name)
	if f.CreatedAt.IsZero() {
		f.CreatedAt = existing.CreatedAt
	}
	f.UpdatedAt = time.Now().UTC()
	s.flags[f.Name] = f
	return nil
}

// Delete removes a flag.
func (s *MemoryStore) Delete(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[name]; !ok {
		return model.ErrNotFound
	}
	delete(s.flags, name)
	return nil
}

// get must be called with the lock held.
func (s *MemoryStore) get(name string) (model.Flag, bool) {
	f, ok := s.flags[name]
	return f, ok
}
