// Package model defines domain types.
package model

import "time"

// Flag is a feature flag.
type Flag struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Enabled        bool      `json:"enabled"`
	RolloutPercent int       `json:"rollout_percent"`
	KillSwitch     bool      `json:"kill_switch"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validate checks invariants.
func (f Flag) Validate() error {
	if f.Name == "" {
		return ErrInvalidName
	}
	if f.RolloutPercent < 0 || f.RolloutPercent > 100 {
		return ErrInvalidRollout
	}
	return nil
}
