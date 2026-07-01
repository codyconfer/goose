package events

import (
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

func TestRollEventCertainFires(t *testing.T) {
	saved := Events
	defer func() { Events = saved }()

	fired := false
	Events = []Event{{
		Key:     "always",
		Trigger: ChanceTrigger{P: 1},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			fired = true
			return Outcome{
				Cmds:  []economy.Command{economy.Earn(10)},
				Notif: notify.Notification{Title: "t", Message: "m", Tone: notify.TonePositive},
			}
		},
	}}

	em := economy.NewMachine()
	es := NewMachine()
	r := rand.New(rand.NewSource(1))
	out, ok := es.Roll(em.Get(), r)
	if !ok || !fired {
		t.Fatal("certain event did not fire")
	}

	em.ApplyWindfall(out.Notif.Title, out.Cmds)
	if out.Notif.Title != "t" || em.Get().Tokens != 10 {
		t.Fatalf("event result/effect wrong: %+v tokens=%v", out, em.Get().Tokens)
	}
}

func TestRollEventNeverFires(t *testing.T) {
	saved := Events
	defer func() { Events = saved }()
	Events = []Event{{Key: "never", Trigger: ChanceTrigger{P: 0}, Apply: func(s economy.State, r *rand.Rand) Outcome { return Outcome{} }}}

	s := economy.NewState()
	es := NewMachine()
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		if _, ok := es.Roll(s, r); ok {
			t.Fatal("zero-chance event fired")
		}
	}
}

func TestRollEventGated(t *testing.T) {
	saved := Events
	defer func() { Events = saved }()
	Events = []Event{{
		Key:     "gated",
		Trigger: ChanceTrigger{P: 1},
		CanFire: func(s economy.State) bool { return s.Tokens >= 100 },
		Apply:   func(s economy.State, r *rand.Rand) Outcome { return Outcome{Notif: notify.Notification{Title: "g"}} },
	}}

	r := rand.New(rand.NewSource(1))
	if _, ok := NewMachine().Roll(economy.NewState(), r); ok {
		t.Fatal("gated event fired while ineligible")
	}
	rich := economy.NewState()
	rich.Tokens = 100
	if _, ok := NewMachine().Roll(rich, r); !ok {
		t.Fatal("gated event did not fire when eligible")
	}
}

func TestRollEventOneShotFiresOnce(t *testing.T) {
	saved := Events
	defer func() { Events = saved }()

	count := 0
	Events = []Event{{
		Key:     "milestone",
		Trigger: LevelTrigger{Level: 1},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			count++
			return Outcome{Notif: notify.Notification{Title: "m"}}
		},
	}}

	s := economy.NewState()
	es := NewMachine()
	r := rand.New(rand.NewSource(1))
	if _, ok := es.Roll(s, r); !ok {
		t.Fatal("one-shot event did not fire the first time")
	}
	for i := 0; i < 100; i++ {
		if _, ok := es.Roll(s, r); ok {
			t.Fatal("one-shot event fired again after being marked")
		}
	}
	if count != 1 {
		t.Fatalf("one-shot applied %d times, want 1", count)
	}
	if !es.Get().HasFired("milestone") {
		t.Fatal("one-shot event was not recorded as fired")
	}
}

func TestLevelTrigger(t *testing.T) {
	tr := LevelTrigger{Level: 3}
	if tr.Repeatable() {
		t.Fatal("level trigger should be one-shot")
	}
	low := economy.NewState()
	if tr.Fires(low, nil) {
		t.Fatal("level trigger fired below its level")
	}
	hi := economy.NewState()
	hi.PeakEggs = economy.LevelThresholds[2]
	if !tr.Fires(hi, nil) {
		t.Fatal("level trigger did not fire at its level")
	}
}

func TestMarketCapTrigger(t *testing.T) {
	tr := MarketCapTrigger{Cap: 1000}
	if tr.Repeatable() {
		t.Fatal("market-cap trigger should be one-shot")
	}
	s := economy.NewState()
	s.EggsLaid = 400
	s.EggsBought = 400
	if tr.Fires(s, nil) {
		t.Fatal("market-cap trigger fired below its cap")
	}
	s.EggsBought = 700
	if !tr.Fires(s, nil) {
		t.Fatal("market-cap trigger did not fire at its cap")
	}
}

func TestEggPriceTrigger(t *testing.T) {
	tr := EggPriceTrigger{High: economy.BasePrice * 1.5, Chance: 1}
	if !tr.Repeatable() {
		t.Fatal("egg-price trigger should be repeatable")
	}
	s := economy.NewState()
	r := rand.New(rand.NewSource(1))

	s.PriceFactor = 1.0
	if tr.Fires(s, r) {
		t.Fatal("egg-price trigger fired inside the band")
	}
	s.PriceFactor = 1.8
	if !tr.Fires(s, r) {
		t.Fatal("egg-price trigger did not fire above High with chance 1")
	}

	never := EggPriceTrigger{High: economy.BasePrice * 1.5, Chance: 0}
	for i := 0; i < 100; i++ {
		if never.Fires(s, r) {
			t.Fatal("zero-chance egg-price trigger fired")
		}
	}
}

func TestReconcile(t *testing.T) {
	saved := Events
	defer func() { Events = saved }()

	applied := false
	Events = []Event{
		{
			Key:     "past_milestone",
			Trigger: LevelTrigger{Level: 1},
			Apply: func(s economy.State, r *rand.Rand) Outcome {
				applied = true
				return Outcome{}
			},
		},
		{
			Key:     "chance",
			Trigger: ChanceTrigger{P: 1},
			Apply:   func(s economy.State, r *rand.Rand) Outcome { return Outcome{} },
		},
	}

	s := economy.NewState()
	es := NewMachine()
	es.Reconcile(s)
	if applied {
		t.Fatal("Reconcile applied an event effect")
	}
	if !es.Get().HasFired("past_milestone") {
		t.Fatal("Reconcile did not mark a satisfied one-shot as fired")
	}
	if es.Get().HasFired("chance") {
		t.Fatal("Reconcile marked a repeatable event as fired")
	}

	r := rand.New(rand.NewSource(1))
	out, ok := es.Roll(s, r)
	if !ok {
		t.Fatal("expected the repeatable chance event to fire")
	}
	if applied {
		t.Fatal("reconciled one-shot fired despite being marked")
	}
	_ = out
}
