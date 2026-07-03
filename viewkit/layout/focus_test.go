package layout

import (
	"slices"
	"testing"
)

func TestNewRingFiltersInteractive(t *testing.T) {
	got := NewRing(
		Focusable{Name: "a", Interactive: true},
		Focusable{Name: "b", Interactive: false},
		Focusable{Name: "c", Interactive: true},
	)
	if !slices.Equal([]string(got), []string{"a", "c"}) {
		t.Fatalf("NewRing = %v, want [a c]", got)
	}
}

func TestRingStepWrapsAndAtClamps(t *testing.T) {
	ring := Ring{"a", "c"}
	if got := ring.Step(0, 1); got != 1 {
		t.Errorf("step forward = %d, want 1", got)
	}
	if got := ring.Step(1, 1); got != 0 {
		t.Errorf("step past the end should wrap to 0, got %d", got)
	}
	if got := ring.Step(0, -1); got != 1 {
		t.Errorf("step before the start should wrap to last, got %d", got)
	}
	if got := ring.At(99); got != "c" {
		t.Errorf("At clamps out-of-range idx, got %q", got)
	}
	if got := (Ring(nil)).At(0); got != "" {
		t.Errorf("At of an empty ring should be empty, got %q", got)
	}
}
