package layout

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/theme"
)

func TestStackFitTallShowsEverything(t *testing.T) {
	secs := []Section{
		{Content: "alpha"},
		{Content: "bravo", MinTier: TierMedium},
		{Content: "charlie", MinTier: TierTall},
	}
	got := StackFit(TierTall, secs...)
	want := Stack("alpha", "bravo", "charlie")
	if got != want {
		t.Fatalf("StackFit(TierTall) = %q, want %q", got, want)
	}
}

func TestStackFitMediumDropsTallOnly(t *testing.T) {
	secs := []Section{
		{Content: "alpha"},
		{Content: "bravo", MinTier: TierMedium},
		{Content: "charlie", MinTier: TierTall},
	}
	got := StackFit(TierMedium, secs...)
	if strings.Contains(got, "charlie") {
		t.Errorf("tall-only section should drop at medium:\n%s", got)
	}
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "bravo") {
		t.Errorf("short and medium sections should survive at medium:\n%s", got)
	}
}

func TestStackFitShortKeepsOnlyEssentials(t *testing.T) {
	secs := []Section{
		{Content: "alpha"},
		{Content: "bravo", MinTier: TierMedium},
		{Content: "charlie", MinTier: TierTall},
	}
	got := StackFit(TierShort, secs...)
	if strings.Contains(got, "bravo") || strings.Contains(got, "charlie") {
		t.Errorf("only short-tier sections should survive at short:\n%s", got)
	}
	if !strings.Contains(got, "alpha") {
		t.Errorf("short-tier (zero value) section must survive:\n%s", got)
	}
}

func TestStackFitSkipsEmptySections(t *testing.T) {
	got := StackFit(TierTall,
		Section{Content: "alpha"},
		Section{Content: "", MinTier: TierMedium},
		Section{Content: "bravo"},
	)
	want := Stack("alpha", "bravo")
	if got != want {
		t.Fatalf("StackFit with empty section = %q, want %q", got, want)
	}
}

func TestTierForHeightBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		height int
		want   Tier
	}{
		{"unknown height falls back to medium", 0, TierMedium},
		{"below minimum is short", theme.MinBodyHeight - 1, TierShort},
		{"exactly minimum is medium", theme.MinBodyHeight, TierMedium},
		{"just below tall is medium", theme.TallBodyHeight - 1, TierMedium},
		{"exactly tall is tall", theme.TallBodyHeight, TierTall},
		{"above tall is tall", theme.TallBodyHeight + 20, TierTall},
	}
	for _, c := range cases {
		if got := TierForHeight(c.height); got != c.want {
			t.Errorf("%s: TierForHeight(%d) = %d, want %d", c.name, c.height, got, c.want)
		}
	}
}

func TestTierRowsAt(t *testing.T) {
	r := TierRows{Short: 3, Medium: 8, Tall: 12}
	if got := r.At(TierShort); got != 3 {
		t.Errorf("At(TierShort) = %d, want 3", got)
	}
	if got := r.At(TierMedium); got != 8 {
		t.Errorf("At(TierMedium) = %d, want 8", got)
	}
	if got := r.At(TierTall); got != 12 {
		t.Errorf("At(TierTall) = %d, want 12", got)
	}
}
