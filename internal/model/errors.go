package model

import "errors"

var (
	ErrInvalidName    = errors.New("flag name is required")
	ErrInvalidRollout = errors.New("rollout_percent must be between 0 and 100")
	ErrNotFound       = errors.New("flag not found")
	ErrAlreadyExists  = errors.New("flag already exists")
)
