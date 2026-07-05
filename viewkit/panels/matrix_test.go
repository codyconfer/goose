package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
)

func TestRainRowCount(t *testing.T) {
	r := NewRain(20, 12, 42)
	if body := r.renderBody(20); len(body) != 12 {
		t.Fatalf("renderBody rows = %d, want 12", len(body))
	}
}

func TestRainRowsAreSingleCell(t *testing.T) {
	r := NewRain(20, 8, 1)
	for i := 0; i < 5; i++ {
		r.Beat()
	}
	for i, row := range r.renderBody(20) {
		if w := ansi.StringWidth(row); w != 20 {
			t.Fatalf("row %d display width = %d, want 20: %q", i, w, stripANSI(row))
		}
	}
}

func TestRainRendersGlyphs(t *testing.T) {
	r := NewRain(30, 10, 3)
	for i := 0; i < 4; i++ {
		r.Beat()
	}
	joined := stripANSI(strings.Join(r.renderBody(30), "\n"))
	if !strings.ContainsAny(joined, string(glyphSet)) {
		t.Fatalf("expected at least one glyph, got only blanks:\n%q", joined)
	}
}

func TestRainMovesOverBeats(t *testing.T) {
	r := NewRain(24, 12, 9)

	col := -1
	for x := range r.cols {
		if r.cols[x].head >= 0 {
			col = x
			break
		}
	}
	if col < 0 {
		t.Fatal("seed produced no active columns")
	}
	oldHead, speed := r.cols[col].head, r.cols[col].speed

	before := stripANSI(strings.Join(r.renderBody(24), "\n"))
	for i := 0; i < speed; i++ {
		r.Beat()
	}
	if got := r.cols[col].head; got != oldHead+1 && got != -1 {
		t.Fatalf("head = %d after %d beats, want %d (or -1 if it exited)", got, speed, oldHead+1)
	}

	after := stripANSI(strings.Join(r.renderBody(24), "\n"))
	if before == after {
		t.Fatal("rain did not change after advancing")
	}
}

func TestRainDeterministic(t *testing.T) {
	a := NewRain(20, 10, 77)
	b := NewRain(20, 10, 77)
	for i := 0; i < 15; i++ {
		a.Beat()
		b.Beat()
	}
	if got, want := strings.Join(a.renderBody(20), "\n"), strings.Join(b.renderBody(20), "\n"); got != want {
		t.Fatal("same seed + same beats produced different output")
	}
}

func TestMatrixWrapsWithTitle(t *testing.T) {
	r := NewRain(20, 6, 5)
	out := stripANSI(Matrix(layout.DefaultFrame(), "MATRIX", r))
	if !strings.Contains(out, "MATRIX") {
		t.Fatalf("panel missing title:\n%s", out)
	}
}
