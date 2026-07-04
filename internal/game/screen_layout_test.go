package game

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func specHasPane(spec layout.ScreenSpec, key string) bool {
	for _, p := range spec.Panes {
		if p.Key == key {
			return true
		}
	}
	return false
}

func tallTradeModel(t *testing.T) Model {
	t.Helper()
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	econ.ScheduleTrade(economy.TxBuyEggs, 100)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	m.width = theme.MinScreenWidth
	m.height = 120
	return m
}

func TestDefaultLayoutShowsLedger(t *testing.T) {
	isolateHome(t)
	m := tallTradeModel(t)
	if !strings.Contains(m.View(), "LEDGER") {
		t.Fatalf("default tall trade desk should show LEDGER:\n%s", m.View())
	}
}

func TestLayoutEditorToggleHidesPaneAndPersists(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig() // reset in-memory config for this temp HOME

	le := newLayoutEditor(&menuScreen{})
	es := le.specs[screenTrade]
	found := false
	for i := range es.panes {
		if es.panes[i].key == "ledger" {
			es.panes[i].on = false
			found = true
		}
	}
	if !found {
		t.Fatal("ledger pane missing from trade editor")
	}
	if err := le.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	if specHasPane(layoutSpec(screenTrade), "ledger") {
		t.Fatal("ledger should be removed from the in-memory trade spec")
	}

	// Reload from disk to prove persistence.
	loadLayoutConfig()
	if specHasPane(layoutSpec(screenTrade), "ledger") {
		t.Fatal("ledger should stay removed after reload from disk")
	}

	if v := tallTradeModel(t).View(); strings.Contains(v, "LEDGER") {
		t.Fatalf("tall trade desk should hide LEDGER after toggle:\n%s", v)
	}
}

func TestLayoutEditorReorderMovesPane(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1 // trade
	if configurableScreens[le.screenIdx] != screenTrade {
		t.Fatalf("expected trade at index 1, got %q", configurableScreens[le.screenIdx])
	}
	le.cursor = 2 // first pane row
	first := le.current().panes[0].key
	second := le.current().panes[1].key
	le.reorder(1) // move first pane down

	if le.current().panes[0].key != second || le.current().panes[1].key != first {
		t.Fatalf("reorder did not swap first two panes: %v", le.current().panes)
	}
	if le.cursor != 3 {
		t.Fatalf("cursor should follow moved pane to row 3, got %d", le.cursor)
	}
}

func TestLayoutEditorChangeLayoutPersists(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1 // trade
	le.cursor = 1    // layout row
	es := le.current()
	// force a known non-default layout key
	es.layoutIdx = layoutIndex(es.layouts, "single")
	if err := le.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	loadLayoutConfig()
	if got := layoutSpec(screenTrade).Layout; got != "single" {
		t.Fatalf("trade layout should persist as single, got %q", got)
	}
}

func TestMenuOpensLayoutEditor(t *testing.T) {
	isolateHome(t)
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth
	m.height = 120
	m.screen = &menuScreen{items: menuItems(nil)}

	m = send(m, key("l"))
	if _, ok := m.screen.(*layoutEditorScreen); !ok {
		t.Fatalf("pressing l should open the layout editor, got %T", m.screen)
	}
	if v := m.View(); !strings.Contains(v, "SCREEN LAYOUT") {
		t.Fatalf("layout editor should render its header:\n%s", v)
	}
}

func TestGameHotkeyOpensLayoutEditor(t *testing.T) {
	isolateHome(t)
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth
	m.height = 120

	if _, ok := m.screen.(*gameScreen); !ok {
		t.Fatalf("expected to start on the game screen, got %T", m.screen)
	}
	m = send(m, key("L"))
	le, ok := m.screen.(*layoutEditorScreen)
	if !ok {
		t.Fatalf("pressing L in game should open the layout editor, got %T", m.screen)
	}
	if _, ok := le.prev.(*gameScreen); !ok {
		t.Fatalf("editor should return to the game screen, prev is %T", le.prev)
	}
	m = send(m, key("esc"))
	if _, ok := m.screen.(*gameScreen); !ok {
		t.Fatalf("esc should return to the game screen, got %T", m.screen)
	}
}
