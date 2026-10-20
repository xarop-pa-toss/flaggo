package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

// PostgresStore implements Store using Postgres.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgres opens a connection pool and returns a store.
func NewPostgres(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

const migrateSQL = `
CREATE TABLE IF NOT EXISTS flags (
    name           TEXT PRIMARY KEY,
    description    TEXT NOT NULL DEFAULT '',
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    rollout_percent INTEGER NOT NULL DEFAULT 0,
    kill_switch    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, migrateSQL)
	return err
}

func (s *PostgresStore) Create(ctx context.Context, f model.Flag) error {
	if err := f.Validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = now
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO flags (name, description, enabled, rollout_percent, kill_switch, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		f.Name, f.Description, f.Enabled, f.RolloutPercent, f.KillSwitch, f.CreatedAt, f.UpdatedAt)
	if err != nil {
		// detect unique violation
		if isUniqueViolation(err) {
			return model.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PostgresStore) Get(ctx context.Context, name string) (model.Flag, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT name, description, enabled, rollout_percent, kill_switch, created_at, updated_at
		FROM flags WHERE name = $1`, name)
	var f model.Flag
	err := row.Scan(&f.Name, &f.Description, &f.Enabled, &f.RolloutPercent, &f.KillSwitch, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Flag{}, model.ErrNotFound
		}
		return model.Flag{}, err
	}
	return f, nil
}

func (s *PostgresStore) List(ctx context.Context) ([]model.Flag, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, description, enabled, rollout_percent, kill_switch, created_at, updated_at
		FROM flags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Flag
	for rows.Next() {
		var f model.Flag
		if err := rows.Scan(&f.Name, &f.Description, &f.Enabled, &f.RolloutPercent, &f.KillSwitch, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *PostgresStore) Update(ctx context.Context, f model.Flag) error {
	if err := f.Validate(); err != nil {
		return err
	}
	f.UpdatedAt = time.Now().UTC()
	tag, err := s.pool.Exec(ctx, `
		UPDATE flags SET description=$2, enabled=$3, rollout_percent=$4, kill_switch=$5, updated_at=$6
		WHERE name=$1`,
		f.Name, f.Description, f.Enabled, f.RolloutPercent, f.KillSwitch, f.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, name string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM flags WHERE name=$1`, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	// pgx wraps pgconn.PgError; check message/code loosely.
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
