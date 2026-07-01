package economy

import (
	"strings"
	"testing"
)

func TestSeizeBest(t *testing.T) {
	m := NewMachine()
	if _, ok := m.s.BestProducer(); ok {
		t.Fatal("found a best producer in an empty flock")
	}
	m.seizeBest()

	m.s.Owned["server"] = 3
	m.s.Owned["gpu"] = 1
	p, ok := m.s.BestProducer()
	if !ok || p.Key != "server" {
		t.Fatalf("best producer %v ok=%v, want server", p.Key, ok)
	}
	m.seizeBest()
	if m.s.Owned["server"] != 2 {
		t.Fatalf("server count=%d, want 2", m.s.Owned["server"])
	}
}

func TestCount(t *testing.T) {
	m := NewMachine()
	m.s.Owned["gpu"] = 4
	if m.s.Count("gpu") != 4 || m.s.Count("missing") != 0 {
		t.Fatalf("Count gpu=%d missing=%d, want 4/0", m.s.Count("gpu"), m.s.Count("missing"))
	}
}

func TestUpgradeUnlockRules(t *testing.T) {
	m := NewMachine()
	click, _ := UpgradeByKey(UpgradeClick)
	if !click.IsUnlocked(m.s) {
		t.Fatal("Enter the Flow State should always be unlocked")
	}
	crier, _ := UpgradeByKey(UpgradeCrier)
	if crier.IsUnlocked(m.s) {
		t.Fatal("Jet Set Huang should be locked before any eggs are laid")
	}
	m.s.Owned["server"] = 1
	if !crier.IsUnlocked(m.s) {
		t.Fatal("Jet Set Huang should unlock once eggs are laid")
	}
	if _, ok := UpgradeByKey("nope"); ok {
		t.Fatal("UpgradeByKey returned ok for an unknown key")
	}
}

func TestFormatNumMagnitudes(t *testing.T) {
	cases := map[float64]string{
		1.5e15: "Q",
		2.5e12: "T",
		3.5e9:  "B",
		4.5e6:  "M",
		5.5e3:  "K",
	}
	for n, suffix := range cases {
		if got := FormatNum(n); !strings.HasSuffix(got, suffix) {
			t.Errorf("FormatNum(%v)=%q, want suffix %q", n, got, suffix)
		}
	}
	if got := FormatNum(250); got != "250" {
		t.Errorf("FormatNum(250)=%q, want 250", got)
	}
	if got := FormatNum(-1.2e6); !strings.HasPrefix(got, "-") {
		t.Errorf("FormatNum negative=%q, want leading -", got)
	}
}
