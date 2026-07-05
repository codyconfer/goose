package layout

import (
	"strings"
	"testing"
)

func groupPane(name, group string, tier Tier) Pane {
	p := flexBoxPane(name)
	p.Group = group
	p.MinTier = tier
	return p
}

func TestFlexSectionsGroupsWithHeaders(t *testing.T) {
	scr := Screen{
		Layout: FlexSections{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			groupPane("alpha", "SPOTDESK", 0),
			groupPane("bravo", "SPOTDESK", 0),
			groupPane("charlie", "DERIVATIVES", 0),
		},
	}
	out := stripANSI(scr.Render(NewFrame(120), TierTall, 0))

	spot := strings.Index(out, "SPOTDESK")
	deriv := strings.Index(out, "DERIVATIVES")
	if spot < 0 || deriv < 0 {
		t.Fatalf("expected both section headers, got:\n%s", out)
	}
	if spot > deriv {
		t.Fatalf("section order not preserved (SPOTDESK should lead):\n%s", out)
	}
	if a, c := strings.Index(out, "alpha"), strings.Index(out, "charlie"); !(spot < a && a < deriv && deriv < c) {
		t.Fatalf("panes not grouped under their headers:\n%s", out)
	}
}

func TestFlexSectionsUngroupedLeadsWithoutHeader(t *testing.T) {
	scr := Screen{
		Layout: FlexSections{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			groupPane("loose", "", 0),
			groupPane("charlie", "DERIVATIVES", 0),
		},
	}
	out := stripANSI(scr.Render(NewFrame(120), TierTall, 0))

	loose := strings.Index(out, "loose")
	deriv := strings.Index(out, "DERIVATIVES")
	if loose < 0 || deriv < 0 {
		t.Fatalf("expected ungrouped pane and section header:\n%s", out)
	}
	if loose > deriv {
		t.Fatalf("ungrouped panes should lead:\n%s", out)
	}
	if !strings.HasPrefix(out, "╭") {
		t.Fatalf("ungrouped leading block should start with its box, not a header:\n%s", out)
	}
}

func TestFlexSectionsSkipsGroupWithNoVisiblePanes(t *testing.T) {
	scr := Screen{
		Layout: FlexSections{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			groupPane("alpha", "SPOTDESK", 0),
			groupPane("charlie", "DERIVATIVES", TierTall),
		},
	}
	out := stripANSI(scr.Render(NewFrame(120), TierShort, 0))

	if strings.Contains(out, "DERIVATIVES") {
		t.Fatalf("group with no visible panes must not render a header:\n%s", out)
	}
	if !strings.Contains(out, "SPOTDESK") || !strings.Contains(out, "alpha") {
		t.Fatalf("visible group should still render:\n%s", out)
	}
	if strings.Contains(out, "charlie") {
		t.Fatalf("tier-hidden pane should not render:\n%s", out)
	}
}

func maxTopBorders(out string) int {
	most := 0
	for _, line := range strings.Split(out, "\n") {
		if n := strings.Count(line, "╭"); n > most {
			most = n
		}
	}
	return most
}

func TestFlexSectionsReflowsColumnsWithinGroup(t *testing.T) {
	scr := Screen{
		Layout: FlexSections{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			groupPane("alpha", "SPOTDESK", 0),
			groupPane("bravo", "SPOTDESK", 0),
			groupPane("charlie", "SPOTDESK", 0),
		},
	}
	wide := scr.Render(NewFrame(120), TierTall, 0)
	if got := maxTopBorders(wide); got != 3 {
		t.Fatalf("wide group should flow into 3 columns, got %d:\n%s", got, wide)
	}
	narrow := scr.Render(NewFrame(50), TierTall, 0)
	if got := maxTopBorders(narrow); got != 1 {
		t.Fatalf("narrow group should collapse to 1 column, got %d:\n%s", got, narrow)
	}
}
