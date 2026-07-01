package economy

import "testing"

func TestApplyOfflineCreditsAndCaps(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 10
	rate := m.s.TokensPerSecond()

	earned := m.ApplyOffline(60)
	if got := earned; got != rate*60 {
		t.Fatalf("offline earned=%v, want %v", got, rate*60)
	}

	m2 := NewMachine()
	m2.s.Owned["server"] = 10
	capped := m2.ApplyOffline(10 * maxOfflineSeconds)
	if capped != rate*float64(maxOfflineSeconds) {
		t.Fatalf("capped offline=%v, want %v", capped, float64(maxOfflineSeconds)*rate)
	}
}

func TestApplyOfflineNoopForNonPositive(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 5
	before := m.s.Tokens
	if got := m.ApplyOffline(0); got != 0 || m.s.Tokens != before {
		t.Fatalf("offline for 0s changed state: earned=%v tokens=%v", got, m.s.Tokens)
	}
}
