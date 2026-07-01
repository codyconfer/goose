package economy

import "testing"

func TestLevelFor(t *testing.T) {
	cases := []struct {
		eggs float64
		want int
	}{
		{0, 1},
		{49, 1},
		{50, 2},
		{749, 2},
		{750, 3},
		{10_000, 4},
		{150_000, 5},
		{2_499_999, 5},
		{2_500_000, 6},
		{40_000_000, 7},
		{600_000_000, 8},
		{9_000_000_000, 9},
		{140_000_000_000, 10},
		{2_000_000_000_000, 11},
		{1e15, 11},
	}
	for _, c := range cases {
		if got := LevelFor(c.eggs); got != c.want {
			t.Errorf("LevelFor(%v)=%d, want %d", c.eggs, got, c.want)
		}
	}
}

func TestNextLevelEggs(t *testing.T) {
	if next, ok := NextLevelEggs(0); !ok || next != 50 {
		t.Fatalf("at level 1: next=%v ok=%v, want 50/true", next, ok)
	}
	if next, ok := NextLevelEggs(800); !ok || next != 10_000 {
		t.Fatalf("at level 3: next=%v ok=%v, want 10000/true", next, ok)
	}
	if next, ok := NextLevelEggs(5_000_000); !ok || next != 40_000_000 {
		t.Fatalf("at level 6: next=%v ok=%v, want 40000000/true", next, ok)
	}
	if _, ok := NextLevelEggs(3_000_000_000_000); ok {
		t.Fatal("at max level: expected ok=false")
	}
}

func TestStateLevelTracksEggs(t *testing.T) {
	m := NewMachine()
	if m.s.Level() != 1 {
		t.Fatalf("fresh flock level=%d, want 1", m.s.Level())
	}
	m.s.Eggs = 750
	if m.s.Level() != 3 {
		t.Fatalf("level at 750 eggs=%d, want 3", m.s.Level())
	}
}

func TestProducerUnlockLevelsMatchThresholds(t *testing.T) {

	for _, p := range Producers {
		if p.UnlockLevel < 1 || p.UnlockLevel > MaxLevel() {
			t.Errorf("%s unlock level %d outside 1..%d", p.Key, p.UnlockLevel, MaxLevel())
		}
	}
}
