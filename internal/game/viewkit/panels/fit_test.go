package panels

import (
	"strings"
	"testing"
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
