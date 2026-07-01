package panels

import (
	"strings"
	"testing"
)

func TestPieEmpty(t *testing.T) {
	out := Pie("Mix", []Datum{{"a", 0}, {"b", -5}}, 20, fnum, "nothing here")
	if !strings.Contains(out, "nothing here") {
		t.Fatalf("empty pie missing placeholder:\n%s", out)
	}
}

func TestPieLegendAndProportions(t *testing.T) {
	data := []Datum{{"cash", 75}, {"eggs", 25}}
	out := Pie("Mix", data, 20, fnum, "")
	for _, want := range []string{"cash", "eggs", "75%", "25%", "■", "█"} {
		if !strings.Contains(out, want) {
			t.Errorf("pie output missing %q:\n%s", want, out)
		}
	}
}

func TestPieBarStaysWithinWidth(t *testing.T) {
	data := []Datum{{"a", 1}, {"b", 1}, {"c", 1}}
	lines := strings.Split(stripANSI(Pie("Mix", data, 10, fnum, "")), "\n")

	var barCells int
	for _, l := range lines {
		if c := strings.Count(l, "█"); c > barCells {
			barCells = c
		}
	}
	if barCells > 10 {
		t.Errorf("stacked bar is %d cells, want <= width 10", barCells)
	}
}
