package panels

import (
	"strings"
	"testing"
)

func TestBarEmpty(t *testing.T) {
	out := Bar("Flow", nil, 20, fnum, "no data")
	if !strings.Contains(out, "no data") {
		t.Fatalf("empty bar missing placeholder:\n%s", out)
	}
}

func TestBarShowsLabelsAndValues(t *testing.T) {
	data := []Datum{{"laying", 10}, {"selling", 4}, {"deficit", -2}}
	out := Bar("Flow", data, 20, fnum, "")
	for _, want := range []string{"laying", "selling", "deficit", "10", "4", "-2", "█"} {
		if !strings.Contains(out, want) {
			t.Errorf("bar output missing %q:\n%s", want, out)
		}
	}
}

func TestBarScalesToLargestMagnitude(t *testing.T) {
	data := []Datum{{"big", 100}, {"small", 10}}
	lines := strings.Split(stripANSI(Bar("F", data, 20, fnum, "")), "\n")
	var big, small int
	for _, l := range lines {
		switch {
		case strings.Contains(l, "big"):
			big = strings.Count(l, "█")
		case strings.Contains(l, "small"):
			small = strings.Count(l, "█")
		}
	}
	if big != 20 {
		t.Errorf("largest bar = %d cells, want full width 20", big)
	}
	if small >= big {
		t.Errorf("smaller value drew %d cells, not fewer than %d", small, big)
	}
}
