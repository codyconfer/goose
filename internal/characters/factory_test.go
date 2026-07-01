package characters

import (
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

func TestPickHonorsWeights(t *testing.T) {
	resolve := pick(
		br(3, flat(outcome(notify.Neutral("heavy", "m")))),
		br(1, flat(outcome(notify.Neutral("light", "m")))),
	)
	counts := map[string]int{}
	r := rand.New(rand.NewSource(7))
	s := economy.NewState()
	for i := 0; i < 6000; i++ {
		counts[resolve(s, r).Notif.Title]++
	}
	if counts["light"] == 0 {
		t.Fatal("low-weight branch never fired")
	}
	if counts["heavy"] < counts["light"]*2 {
		t.Fatalf("weighting off: heavy=%d light=%d", counts["heavy"], counts["light"])
	}
}

func TestChanceSplitsTwoWays(t *testing.T) {
	resolve := chance(0.5,
		flat(outcome(notify.Positive("win", "m"))),
		flat(outcome(notify.Negative("lose", "m"))),
	)
	seen := map[string]bool{}
	r := rand.New(rand.NewSource(1))
	s := economy.NewState()
	for i := 0; i < 200; i++ {
		seen[resolve(s, r).Notif.Title] = true
	}
	if !seen["win"] || !seen["lose"] {
		t.Fatalf("chance should reach both branches, saw %v", seen)
	}
}

func TestAffordableClampsToBalance(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 100
	if got := affordable(s, 250); got != 100 {
		t.Fatalf("cost above balance should clamp to tokens, got %v", got)
	}
	if got := affordable(s, 40); got != 40 {
		t.Fatalf("affordable cost should pass through, got %v", got)
	}
	if got := affordable(s, -5); got != 0 {
		t.Fatalf("negative cost should clamp to zero, got %v", got)
	}
}
