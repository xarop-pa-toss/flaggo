// Package eval evaluates flags in memory.
package eval

import (
	"hash/fnv"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

// Evaluate returns whether a flag is on for the given user key.
// Kill switch overrides everything. Disabled flags are off.
// Rollout percentage is applied deterministically by user key.
func Evaluate(f model.Flag, userKey string) bool {
	if f.KillSwitch {
		return false
	}
	if !f.Enabled {
		return false
	}
	if f.RolloutPercent >= 100 {
		return true
	}
	if f.RolloutPercent <= 0 {
		return false
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(f.Name))
	_, _ = h.Write([]byte{':'})
	_, _ = h.Write([]byte(userKey))
	bucket := int(h.Sum32() % 100)
	return bucket < f.RolloutPercent
}
