package characters

import (
	"math"
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
)

type spawnFn func(economy.State, *rand.Rand) Character

func TestNewcomerOptionsResolveAllBranches(t *testing.T) {
	cases := []struct {
		name  string
		spawn spawnFn
		typ   Type
	}{
		{"wook", NewWook, Wook},
		{"booster", NewBooster, Booster},
		{"politician", NewPolitician, Politician},
		{"engineer", NewEngineer, Engineer},
		{"middleclass", NewMiddleClass, MiddleClass},
		{"analyst", NewAnalyst, Analyst},
		{"shortseller", NewShortSeller, ShortSeller},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for seed := int64(0); seed < 120; seed++ {
				m := economy.FromState(economy.State{
					Tokens:      50000,
					TotalEarned: 50000,
					Owned:       map[string]int{"server": 3},
					Consumers:   10,
				})
				r := rand.New(rand.NewSource(seed))
				ch := tc.spawn(m.Get(), r)
				if ch.Type != tc.typ {
					t.Fatalf("seed %d: type=%v want %v", seed, ch.Type, tc.typ)
				}
				if ch.Name == "" || ch.Pitch == "" || len(ch.Options) == 0 {
					t.Fatalf("seed %d: malformed character %+v", seed, ch)
				}
				for opt := range ch.Options {
					out := ch.Options[opt].Resolve(m.Get(), r)
					if out.Notif.Title == "" || out.Notif.Message == "" {
						t.Fatalf("%s seed %d opt %d: empty result", tc.name, seed, opt)
					}
					m.ApplyWindfall(out.Notif.Title, out.Cmds)
					if math.IsNaN(m.Get().Tokens) || m.Get().Tokens < 0 {
						t.Fatalf("%s seed %d opt %d: invalid tokens %v", tc.name, seed, opt, m.Get().Tokens)
					}
				}
			}
		})
	}
}

func TestNewcomerEligibility(t *testing.T) {
	fresh := economy.FromState(economy.State{}).Get()
	if WookEligible(fresh) {
		t.Fatal("a brand-new flock should not attract a wook")
	}
	if BoosterEligible(fresh) {
		t.Fatal("a brand-new flock should not attract Jet Set Huang")
	}

	established := economy.FromState(economy.State{TotalEarned: 50000}).Get()
	if !WookEligible(established) {
		t.Fatal("an established flock should attract a wook")
	}
	if !BoosterEligible(established) {
		t.Fatal("a rich flock should attract Jet Set Huang")
	}
	if PoliticianEligible(fresh) {
		t.Fatal("a brand-new flock should not attract a regulator")
	}
	if !PoliticianEligible(established) {
		t.Fatal("an established flock should attract a regulator")
	}
	if EngineerEligible(fresh) {
		t.Fatal("a brand-new flock has no bubble for an engineer to pop")
	}
	if !EngineerEligible(established) {
		t.Fatal("an established flock should attract an engineer")
	}
	if MiddleClassEligible(fresh) {
		t.Fatal("a brand-new flock should not attract a delegation")
	}
	if !MiddleClassEligible(established) {
		t.Fatal("an established flock should attract the squeezed middle")
	}
	if AnalystEligible(fresh) {
		t.Fatal("a brand-new flock has no story for an analyst to sell")
	}
	if !AnalystEligible(established) {
		t.Fatal("an established flock should attract a sell-side analyst")
	}
	if ShortSellerEligible(fresh) {
		t.Fatal("a brand-new flock is too small to bother shorting")
	}
	if !ShortSellerEligible(established) {
		t.Fatal("a bubbly flock should attract a short-seller")
	}
}
