// Package store defines persistence for flags.
package store

import (
	"context"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

// Store is the persistence interface for flags.
type Store interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, f model.Flag) error
	Get(ctx context.Context, name string) (model.Flag, error)
	List(ctx context.Context) ([]model.Flag, error)
	Update(ctx context.Context, f model.Flag) error
	Delete(ctx context.Context, name string) error
	Close() error
}
