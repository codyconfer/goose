package events

import (
	"math"
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

func richState() economy.State {
	return economy.FromState(economy.State{
		Tokens:      100_000,
		TotalEarned: 5_000_000,
		PerClick:    10,
		Owned:       map[string]int{"server": 5, "rack": 3, "datacenter": 2},
		Consumers:   50,
		Eggs:        5000,
		EggsLaid:    5000,
		PriceFactor: 1.4,
	}).Get()
}

func TestEventApplyProducesValidOutcomes(t *testing.T) {
	for i := range Events {
		e := Events[i]
		for seed := int64(0); seed < 80; seed++ {
			m := economy.FromState(richState())
			r := rand.New(rand.NewSource(seed))
			out := e.Apply(m.Get(), r)
			if out.Notif.Title == "" || out.Notif.Message == "" {
				t.Fatalf("event %s seed %d: empty notification", e.Key, seed)
			}
			m.ApplyWindfall(out.Notif.Title, out.Cmds)
			got := m.Get().Tokens
			if math.IsNaN(got) || got < 0 {
				t.Fatalf("event %s seed %d: invalid tokens %v", e.Key, seed, got)
			}
		}
	}
}

func TestPickHonorsWeights(t *testing.T) {
	tally := func(a apply) map[string]int {
		counts := map[string]int{}
		r := rand.New(rand.NewSource(7))
		s := economy.NewState()
		for i := 0; i < 6000; i++ {
			counts[a(s, r).Notif.Title]++
		}
		return counts
	}

	heavy := pick(
		br(3, func(economy.State, *rand.Rand) Outcome { return outcome(notify.Neutral("heavy", "m")) }),
		br(1, func(economy.State, *rand.Rand) Outcome { return outcome(notify.Neutral("light", "m")) }),
	)
	c := tally(heavy)
	if c["light"] == 0 {
		t.Fatal("low-weight branch never fired")
	}
	if c["heavy"] < c["light"]*2 {
		t.Fatalf("weighting off: heavy=%d light=%d", c["heavy"], c["light"])
	}
}

func TestPickFallsBackWhenNoWeight(t *testing.T) {
	a := pick(
		br(0, func(economy.State, *rand.Rand) Outcome { return outcome(notify.Neutral("zero", "m")) }),
		br(-1, func(economy.State, *rand.Rand) Outcome { return outcome(notify.Neutral("neg", "m")) }),
		br(0, func(economy.State, *rand.Rand) Outcome { return outcome(notify.Neutral("last", "m")) }),
	)
	if got := a(economy.NewState(), rand.New(rand.NewSource(1))).Notif.Title; got != "last" {
		t.Fatalf("no-weight pick should fall back to the last branch, got %q", got)
	}
}
