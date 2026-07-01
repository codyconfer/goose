package characters

import (
	"math"
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

func TestNewVCShape(t *testing.T) {
	m := economy.FromState(economy.State{TotalEarned: 5000})
	r := rand.New(rand.NewSource(1))
	ch := NewVC(m.Get(), r)

	if ch.Type != VC {
		t.Fatalf("character type=%v, want VC", ch.Type)
	}
	if ch.Name == "" || ch.Pitch == "" {
		t.Fatal("character missing name/pitch")
	}
	if ch.Stakes <= 0 {
		t.Fatalf("stakes=%v, want > 0", ch.Stakes)
	}
	if len(ch.Options) != 3 {
		t.Fatalf("got %d options, want 3", len(ch.Options))
	}
	for _, o := range ch.Options {
		if o.Label == "" || o.Resolve == nil {
			t.Fatalf("bad option: %+v", o)
		}
	}
}

func TestVCEligibility(t *testing.T) {
	m := economy.FromState(economy.State{})
	if VCEligible(m.Get()) {
		t.Fatal("a brand-new flock should not attract a VC")
	}
	m = economy.FromState(economy.State{TotalEarned: 2000})
	if !VCEligible(m.Get()) {
		t.Fatal("an established flock should attract a VC")
	}
}

func TestTakeTheMoneyCreditsInvestment(t *testing.T) {
	m := economy.FromState(economy.State{TotalEarned: 5000})
	r := rand.New(rand.NewSource(1))
	ch := NewVC(m.Get(), r)

	before := m.Get().Tokens
	out := ch.Options[0].Resolve(m.Get(), r)
	if out.Notif.Title == "" {
		t.Fatal("resolve returned empty result")
	}
	m.ApplyWindfall(out.Notif.Title, out.Cmds)
	if m.Get().Tokens == before && out.Notif.Tone == notify.TonePositive {
		t.Fatal("clean deal did not change tokens")
	}
}

func TestVCOptionsResolveAllBranches(t *testing.T) {
	for seed := int64(0); seed < 120; seed++ {
		for opt := 0; opt < 3; opt++ {
			m := economy.FromState(economy.State{
				Tokens:      5000,
				TotalEarned: 5000,
				Owned:       map[string]int{"server": 3},
				Consumers:   10,
			})
			r := rand.New(rand.NewSource(seed))
			ch := NewVC(m.Get(), r)
			out := ch.Options[opt].Resolve(m.Get(), r)
			if out.Notif.Title == "" || out.Notif.Message == "" {
				t.Fatalf("seed %d opt %d produced an empty result", seed, opt)
			}
			m.ApplyWindfall(out.Notif.Title, out.Cmds)
			if math.IsNaN(m.Get().Tokens) || m.Get().Tokens < 0 {
				t.Fatalf("seed %d opt %d left tokens invalid: %v", seed, opt, m.Get().Tokens)
			}
		}
	}
}

func TestRoll(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	empty := economy.FromState(economy.State{})
	if _, ok := Roll(empty.Get(), r); ok {
		t.Fatal("a brand-new flock attracted a character")
	}

	m := economy.FromState(economy.State{TotalEarned: 5000})
	got := false
	for i := 0; i < 5000; i++ {
		if ch, ok := Roll(m.Get(), r); ok {
			switch ch.Type {
			case VC, Wook, Booster, Politician, Engineer, MiddleClass, Analyst, ShortSeller:
			default:
				t.Fatalf("rolled unexpected character type %v", ch.Type)
			}
			got = true
			break
		}
	}
	if !got {
		t.Fatal("Roll never produced a character over many beats")
	}
}

func TestHeadline(t *testing.T) {
	if VC.Headline() == "" {
		t.Fatal("VC headline is empty")
	}
	if Type(99).Headline() == "" {
		t.Fatal("unknown type should still have a fallback headline")
	}
}
