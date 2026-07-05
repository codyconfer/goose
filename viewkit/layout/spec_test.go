package layout

import (
	"encoding/json"
	"strings"
	"testing"
)

type testCtx struct {
	showBonus bool
}

func testRegistry() *Registry[testCtx] {
	r := NewRegistry[testCtx]()
	r.Pane("status", "Status", func(testCtx) (Pane, bool) {
		return fixedPane("status", false, nil), true
	})
	r.Pane("feed", "Feed", func(c testCtx) (Pane, bool) {
		return fixedPane("feed", true, nil), true
	})
	r.Pane("bonus", "Bonus", func(c testCtx) (Pane, bool) {
		return fixedPane("bonus", false, nil), c.showBonus
	})
	return r
}

func TestBuildScreenMatchesHandBuilt(t *testing.T) {
	r := testRegistry()
	spec := ScreenSpec{
		Layout: "flex-columns",
		Panes:  []PaneRef{{Key: "status"}, {Key: "feed"}},
	}
	scr, err := BuildScreen(spec, testCtx{}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	got := scr.Render(Frame{Width: 80, Height: 6}, TierTall, 0)

	want := Screen{
		Layout: FlexColumns{},
		Panes:  []Pane{fixedPane("status", false, nil), fixedPane("feed", true, nil)},
	}.Render(Frame{Width: 80, Height: 6}, TierTall, 0)

	if got != want {
		t.Fatalf("built screen render mismatch:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestBuildScreenSkipsUnavailablePanes(t *testing.T) {
	r := testRegistry()
	spec := ScreenSpec{
		Layout: "single",
		Panes:  []PaneRef{{Key: "status"}, {Key: "bonus"}, {Key: "feed"}},
	}

	off, err := BuildScreen(spec, testCtx{showBonus: false}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	if len(off.Panes) != 2 {
		t.Fatalf("bonus off: got %d panes, want 2", len(off.Panes))
	}
	if strings.Contains(off.Render(Frame{Width: 40}, TierTall, 0), "bonus") {
		t.Fatalf("unavailable bonus pane should not render")
	}

	on, err := BuildScreen(spec, testCtx{showBonus: true}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	if len(on.Panes) != 3 {
		t.Fatalf("bonus on: got %d panes, want 3", len(on.Panes))
	}
}

func TestBuildScreenUnknownKeysError(t *testing.T) {
	r := testRegistry()

	if _, err := BuildScreen(ScreenSpec{Layout: "nope", Panes: []PaneRef{{Key: "status"}}}, testCtx{}, r); err == nil {
		t.Fatalf("unknown layout should error")
	}
	if _, err := BuildScreen(ScreenSpec{Layout: "single", Panes: []PaneRef{{Key: "ghost"}}}, testCtx{}, r); err == nil {
		t.Fatalf("unknown pane should error")
	}
}

func TestBuildScreenEmptySpec(t *testing.T) {
	r := testRegistry()
	scr, err := BuildScreen(ScreenSpec{Layout: "single"}, testCtx{}, r)
	if err != nil {
		t.Fatalf("empty spec should not error: %v", err)
	}
	if len(scr.Panes) != 0 {
		t.Fatalf("empty spec should yield 0 panes, got %d", len(scr.Panes))
	}
}

func TestBuildScreenAppliesOverrides(t *testing.T) {
	r := testRegistry()
	tier := TierTall
	spec := ScreenSpec{
		Layout: "grid",
		Panes: []PaneRef{
			{Key: "status", Pos: &GridPos{Col: 1, Row: 0}, MinTier: &tier},
		},
	}
	scr, err := BuildScreen(spec, testCtx{}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	p := scr.Panes[0]
	if p.Pos == nil || p.Pos.Col != 1 {
		t.Fatalf("Pos override not applied: %+v", p.Pos)
	}
	if p.MinTier != TierTall {
		t.Fatalf("MinTier override = %d, want %d", p.MinTier, TierTall)
	}
}

func TestBuildScreenRingOrder(t *testing.T) {
	r := testRegistry()
	spec := ScreenSpec{Layout: "single", Panes: []PaneRef{{Key: "status"}, {Key: "feed"}}}
	scr, err := BuildScreen(spec, testCtx{}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	ring := scr.Ring()
	if got := ring.At(0); got != "feed" {
		t.Fatalf("ring.At(0) = %q, want feed", got)
	}
}

func TestScreenSpecJSONRoundTrip(t *testing.T) {
	spec := ScreenSpec{
		Layout:       "flex-columns",
		LayoutParams: Params{"minWidth": 30, "maxCols": 2},
		Panes:        []PaneRef{{Key: "status"}, {Key: "feed", MinTier: tierPtr(TierMedium)}},
	}
	blob, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back ScreenSpec
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Layout != "flex-columns" || len(back.Panes) != 2 {
		t.Fatalf("round-trip lost fields: %+v", back)
	}
	if back.LayoutParams.Int("minWidth", 0) != 30 || back.LayoutParams.Int("maxCols", 0) != 2 {
		t.Fatalf("Params.Int failed on decoded JSON: %+v", back.LayoutParams)
	}
	if back.Panes[1].MinTier == nil || *back.Panes[1].MinTier != TierMedium {
		t.Fatalf("MinTier lost in round-trip: %+v", back.Panes[1])
	}
}

func TestParamsInt(t *testing.T) {
	p := Params{"a": 5, "b": float64(7), "c": "x"}
	if p.Int("a", 0) != 5 {
		t.Fatalf("int value")
	}
	if p.Int("b", 0) != 7 {
		t.Fatalf("float64 value")
	}
	if p.Int("c", 99) != 99 {
		t.Fatalf("non-number should fall back")
	}
	if p.Int("missing", 42) != 42 {
		t.Fatalf("missing key should fall back")
	}
}

func tierPtr(t Tier) *Tier { return &t }

func TestBuildScreenAppliesSlim(t *testing.T) {
	r := testRegistry()
	spec := ScreenSpec{
		Layout: "grid",
		Panes:  []PaneRef{{Key: "status", Slim: true}, {Key: "feed"}},
	}
	scr, err := BuildScreen(spec, testCtx{}, r)
	if err != nil {
		t.Fatalf("BuildScreen: %v", err)
	}
	if !scr.Panes[0].Slim {
		t.Fatalf("slim flag from PaneRef should land on the pane")
	}
	if scr.Panes[1].Slim {
		t.Fatalf("non-slim ref should leave pane.Slim false")
	}
}

func TestScreenSpecSlimJSONRoundTrip(t *testing.T) {
	spec := ScreenSpec{
		Layout:       "grid",
		LayoutParams: Params{"cols": 4, "rows": 4},
		Panes:        []PaneRef{{Key: "status", Slim: true}, {Key: "feed"}},
	}
	blob, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back ScreenSpec
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !back.Panes[0].Slim || back.Panes[1].Slim {
		t.Fatalf("slim flag lost in round-trip: %+v", back.Panes)
	}
	if back.LayoutParams.Int("cols", 0) != 4 || back.LayoutParams.Int("rows", 0) != 4 {
		t.Fatalf("grid params lost in round-trip: %+v", back.LayoutParams)
	}
}
