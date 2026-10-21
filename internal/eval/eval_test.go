package eval

import (
	"strconv"
	"testing"

	"github.com/xarop-pa-toss/flaggo/internal/model"
)

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name string
		flag model.Flag
		user string
		want bool
	}{
		{"disabled", model.Flag{Name: "f", Enabled: false, RolloutPercent: 100}, "u", false},
		{"kill switch", model.Flag{Name: "f", Enabled: true, RolloutPercent: 100, KillSwitch: true}, "u", false},
		{"0 percent", model.Flag{Name: "f", Enabled: true, RolloutPercent: 0}, "u", false},
		{"100 percent", model.Flag{Name: "f", Enabled: true, RolloutPercent: 100}, "u", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Evaluate(c.flag, c.user)
			if got != c.want {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestEvaluateDeterministic(t *testing.T) {
	f := model.Flag{Name: "f", Enabled: true, RolloutPercent: 50}
	a := Evaluate(f, "alice")
	b := Evaluate(f, "alice")
	if a != b {
		t.Fatalf("non-deterministic: %v vs %v", a, b)
	}
}

func TestEvaluateDistribution(t *testing.T) {
	f := model.Flag{Name: "f", Enabled: true, RolloutPercent: 30}
	on := 0
	n := 10000
	for i := 0; i < n; i++ {
		if Evaluate(f, strconv.Itoa(i)) {
			on++
		}
	}
	rate := float64(on) / float64(n)
	if rate < 0.25 || rate > 0.35 {
		t.Fatalf("rollout rate %.3f out of expected ~0.30", rate)
	}
}
