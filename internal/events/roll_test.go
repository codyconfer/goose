package events

import (
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/world"
)

func TestRollMarksOneShotEvents(t *testing.T) {
	catalog := []world.Event{{
		Key:     "milestone",
		Trigger: world.Trigger{Type: "level", Level: 1},
		Outcomes: []world.Outcome{{
			Weight:  1,
			Tone:    "positive",
			Title:   "t",
			Message: "m",
			Effects: []world.Effect{{Type: "earn", Value: world.Expr{Value: 10}}},
		}},
	}}

	es := NewMachine()
	out, ok := es.Roll(catalog, economy.NewState(), rand.New(rand.NewSource(1)))
	if !ok {
		t.Fatal("expected one-shot event to fire")
	}
	if out.Notif.Title != "t" {
		t.Fatalf("title=%q, want t", out.Notif.Title)
	}
	if !es.Get().HasFired("milestone") {
		t.Fatal("expected event to be marked fired")
	}
	if _, ok := es.Roll(catalog, economy.NewState(), rand.New(rand.NewSource(1))); ok {
		t.Fatal("expected one-shot event to stay suppressed")
	}
}

func TestReconcileMarksPastMilestones(t *testing.T) {
	catalog := []world.Event{{
		Key:     "late_milestone",
		Trigger: world.Trigger{Type: "level", Level: 2},
		Outcomes: []world.Outcome{{
			Weight:  1,
			Tone:    "positive",
			Title:   "t",
			Message: "m",
		}},
	}}

	s := economy.NewState()
	s.Eggs = 1_000
	es := NewMachine()
	es.Reconcile(catalog, s)
	if !es.Get().HasFired("late_milestone") {
		t.Fatal("expected reconcile to mark the milestone fired")
	}
}
