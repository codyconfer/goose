package layout

import (
	"strings"
	"testing"
)

func markerPane(name string, interactive bool, tier Tier) Pane {
	return Pane{
		Name:        name,
		Interactive: interactive,
		MinTier:     tier,
		Render: func(f Frame) string {
			if f.Focused {
				return name + ":focused"
			}
			return name
		},
	}
}

func TestPaneRingSkipsNonInteractive(t *testing.T) {
	ring := PaneRing([]Pane{
		markerPane("a", true, TierShort),
		{Name: "title"},
		markerPane("b", false, TierShort),
		markerPane("c", true, TierShort),
	})
	if len(ring) != 2 || ring[0] != "a" || ring[1] != "c" {
		t.Fatalf("ring = %v, want [a c]", ring)
	}
}

func TestSingleColumnFocusesRingSelection(t *testing.T) {
	scr := Screen{Panes: []Pane{
		markerPane("a", true, TierShort),
		markerPane("b", true, TierShort),
	}}
	out := scr.Render(NewFrame(80), TierTall, 1)
	if !strings.Contains(out, "b:focused") {
		t.Fatalf("focused pane b not rendered focused:\n%s", out)
	}
	if strings.Contains(out, "a:focused") {
		t.Fatalf("unfocused pane a rendered focused:\n%s", out)
	}
}

func TestSingleColumnHidesPanesBelowTier(t *testing.T) {
	scr := Screen{Panes: []Pane{
		markerPane("always", false, TierShort),
		markerPane("tallonly", false, TierTall),
	}}
	short := scr.Render(NewFrame(80), TierShort, 0)
	if strings.Contains(short, "tallonly") {
		t.Fatalf("tall-only pane leaked into short tier:\n%s", short)
	}
	if !strings.Contains(short, "always") {
		t.Fatalf("always pane missing at short tier:\n%s", short)
	}
	tall := scr.Render(NewFrame(80), TierTall, 0)
	if !strings.Contains(tall, "tallonly") {
		t.Fatalf("tall-only pane missing at tall tier:\n%s", tall)
	}
}
