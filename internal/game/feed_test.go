package game

import (
	"fmt"
	"slices"
	"testing"
)

func TestFeedRetainsHistoryUpToCap(t *testing.T) {
	var f feed
	for i := 0; i < feedHistory+5; i++ {
		f.push(fmt.Sprintf("line %d", i))
	}
	got := f.lines()
	if len(got) != feedHistory {
		t.Fatalf("feed kept %d lines, want cap %d", len(got), feedHistory)
	}
	if got[0] != "line 5" {
		t.Fatalf("oldest retained line = %q, want %q", got[0], "line 5")
	}
	if got[len(got)-1] != fmt.Sprintf("line %d", feedHistory+4) {
		t.Fatalf("newest line = %q", got[len(got)-1])
	}
}

func TestFeedPersistsWithoutTimer(t *testing.T) {
	var f feed
	f.push("only")
	if got := f.lines(); !slices.Equal(got, []string{"only"}) {
		t.Fatalf("feed=%v, want [only] to persist until rolled off", got)
	}
}

func TestFeedDedupesConsecutiveWithCount(t *testing.T) {
	var f feed
	f.push("bought a server")
	f.push("bought a server")
	f.push("bought a server")
	if got := f.lines(); !slices.Equal(got, []string{"bought a server (x3)"}) {
		t.Fatalf("feed=%v, want a single line tagged (x3)", got)
	}

	f.push("goose escaped")
	f.push("bought a server")
	if got := f.lines(); !slices.Equal(got, []string{"bought a server (x3)", "goose escaped", "bought a server"}) {
		t.Fatalf("feed=%v, want non-consecutive repeat to stay separate", got)
	}
}

func TestFeedIgnoresEmpty(t *testing.T) {
	var f feed
	f.push("")
	if f.active() {
		t.Fatalf("feed accepted empty item: %v", f.lines())
	}
}
