package economy

import "testing"

func frozenMachine(seconds float64) *Machine {
	m := FromState(State{Tokens: 1000, Owned: map[string]int{"server": 5}})
	m.apply([]Command{Freeze(seconds, "the eggs are too round")})
	return m
}

func TestFreezeBlocksProductionAndTap(t *testing.T) {
	m := frozenMachine(30)
	if !m.Get().Frozen() {
		t.Fatal("machine should be frozen after Freeze command")
	}

	before := m.Get().Tokens
	m.Produce(5)
	m.Tap()
	if m.Get().Tokens != before {
		t.Fatalf("frozen flock earned tokens: before=%v after=%v", before, m.Get().Tokens)
	}
}

func TestTickFreezeCountsDownAndExpires(t *testing.T) {
	m := frozenMachine(10)

	if !m.TickFreeze(4) {
		t.Fatal("TickFreeze should report frozen while time remains")
	}
	if got := m.Get().FreezeSeconds; got != 6 {
		t.Fatalf("FreezeSeconds=%v, want 6", got)
	}

	if !m.TickFreeze(100) {
		t.Fatal("the final frozen beat should still report frozen")
	}
	s := m.Get()
	if s.Frozen() || s.FreezeSeconds != 0 || s.FreezeReason != "" {
		t.Fatalf("freeze did not clear: %+v", s)
	}
	if m.TickFreeze(1) {
		t.Fatal("TickFreeze should report not-frozen once the shutdown ends")
	}

	before := m.Get().Tokens
	m.Produce(1)
	if m.Get().Tokens <= before {
		t.Fatal("production did not resume after the shutdown ended")
	}
}

func TestFreezeLongerWins(t *testing.T) {
	m := frozenMachine(10)
	m.apply([]Command{Freeze(5, "shorter")})
	if got := m.Get().FreezeSeconds; got != 10 {
		t.Fatalf("a shorter shutdown overrode a longer one: %v", got)
	}
	m.apply([]Command{Freeze(40, "longer")})
	if got := m.Get().FreezeSeconds; got != 40 {
		t.Fatalf("a longer shutdown did not extend: %v", got)
	}
}

func TestApplyOfflineConsumesFreezeFirst(t *testing.T) {
	m := frozenMachine(20)
	rate := m.Get().TokensPerSecond()
	if rate <= 0 {
		t.Fatal("test needs a producing flock")
	}

	earned := m.ApplyOffline(60)
	if want := rate * 40; earned != want {
		t.Fatalf("offline earnings=%v, want %v (freeze should eat 20s)", earned, want)
	}
	if m.Get().Frozen() {
		t.Fatal("freeze should be cleared after a long absence")
	}
}
