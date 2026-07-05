package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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
	loadLayoutConfig()

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
	le.screenIdx = 1
	if configurableScreens[le.screenIdx] != screenTrade {
		t.Fatalf("expected trade at index 1, got %q", configurableScreens[le.screenIdx])
	}
	base := le.paneBase()
	le.cursor = base
	first := le.current().panes[0].key
	second := le.current().panes[1].key
	le.reorder(1)

	if le.current().panes[0].key != second || le.current().panes[1].key != first {
		t.Fatalf("reorder did not swap first two panes: %v", le.current().panes)
	}
	if le.cursor != base+1 {
		t.Fatalf("cursor should follow moved pane to row %d, got %d", base+1, le.cursor)
	}
}

func TestLayoutEditorChangeLayoutPersists(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1
	le.cursor = 1
	es := le.current()
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

func TestLayoutEditorShowsReorderAffordance(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth
	m.height = 120
	m.screen = le

	if v := le.view(&m); strings.Contains(v, "move panel") {
		t.Fatalf("reorder hint should not show on the screen-selector row:\n%s", v)
	}

	base := le.paneBase()
	le.cursor = base + 1
	v := le.view(&m)
	if !strings.Contains(v, "move panel") {
		t.Fatalf("panel row should show the 'move panel' hint:\n%s", v)
	}
	if !strings.Contains(v, "▾") || !strings.Contains(v, "▴") {
		t.Fatalf("a middle panel row should show both up and down arrows:\n%s", v)
	}

	le.cursor = base
	v = le.view(&m)
	if strings.Contains(v, "▴") {
		t.Fatalf("the first panel row should not offer a move-up arrow:\n%s", v)
	}
	if !strings.Contains(v, "▾") {
		t.Fatalf("the first panel row should still offer a move-down arrow:\n%s", v)
	}
}

func TestLayoutEditorFlexParamsPersist(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1
	es := le.current()

	es.layoutIdx = layoutIndex(es.layouts, "flex-columns")
	if le.currentLayoutKey() != "flex-columns" {
		t.Fatalf("expected flex-columns layout, got %q", le.currentLayoutKey())
	}

	specs := le.paramSpecs()
	if len(specs) != 2 || specs[0].key != "minWidth" || specs[1].key != "maxCols" {
		t.Fatalf("flex should expose minWidth+maxCols params, got %+v", specs)
	}

	// maxCols is the second param row (index 1) → editor row paneBase-relative 2+1.
	le.cursor = 3
	for i := 0; i < 10; i++ {
		le.changeRow(1)
	}
	if got := le.paramValue(specs[1]); got != 4 {
		t.Fatalf("maxCols should clamp to 4, got %d", got)
	}

	if err := le.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	loadLayoutConfig()
	spec := layoutSpec(screenTrade)
	if spec.Layout != "flex-columns" {
		t.Fatalf("layout should persist as flex-columns, got %q", spec.Layout)
	}
	if spec.LayoutParams.Int("maxCols", 0) != 4 {
		t.Fatalf("maxCols should persist as 4, got %d", spec.LayoutParams.Int("maxCols", 0))
	}
}

func TestLayoutEditorThemePickerSwitchesAndPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer theme.Use(theme.Default())

	m := New(economy.NewMachine(), events.NewMachine(), 0)
	le := newLayoutEditor(&menuScreen{})
	m.screen = le

	if !strings.Contains(le.view(&m), "Theme") {
		t.Fatalf("layout editor view missing Theme row:\n%s", le.view(&m))
	}

	le.cursor = le.themePos()
	startAccent := theme.Cur().Accent.GetForeground()
	le.handleKey(&m, tea.KeyMsg{Type: tea.KeyRight})

	if theme.Cur().Accent.GetForeground() == startAccent {
		t.Fatal("stepping theme did not change the active theme")
	}
	wantKey := theme.Keys()[1]
	if theme.Cur().Accent.GetForeground() != mustAccent(t, wantKey) {
		t.Fatalf("active theme = %v, want %s", theme.Cur().Accent.GetForeground(), wantKey)
	}

	data, err := os.ReadFile(filepath.Join(home, ".goose", "theme.json"))
	if err != nil {
		t.Fatalf("theme.json not written: %v", err)
	}
	if !strings.Contains(string(data), wantKey) {
		t.Fatalf("theme.json missing %q: %s", wantKey, data)
	}
}

func TestLayoutEditorShowsLayoutDisplayName(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1
	le.current().layoutIdx = layoutIndex(le.current().layouts, "flex-rows")
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth
	m.height = 120

	if v := le.view(&m); !strings.Contains(v, "Flex Rows") {
		t.Fatalf("layout selector should show the display name 'Flex Rows':\n%s", v)
	}
}

func TestLayoutEditorGridParamsAndSlimPersist(t *testing.T) {
	isolateHome(t)
	loadLayoutConfig()

	le := newLayoutEditor(&menuScreen{})
	le.screenIdx = 1
	es := le.current()

	es.layoutIdx = layoutIndex(es.layouts, "grid")
	if le.currentLayoutKey() != "grid" {
		t.Fatalf("expected grid layout, got %q", le.currentLayoutKey())
	}

	specs := le.paramSpecs()
	if len(specs) != 2 || specs[0].key != "cols" || specs[1].key != "rows" {
		t.Fatalf("grid should expose cols+rows params, got %+v", specs)
	}

	le.cursor = 2
	for i := 0; i < 10; i++ {
		le.changeRow(1)
	}
	if got := le.paramValue(specs[0]); got != 4 {
		t.Fatalf("cols should clamp to 4, got %d", got)
	}

	le.cursor = le.paneBase()
	if _, ok := le.paneIndex(); !ok {
		t.Fatal("expected a panel row at paneBase")
	}
	le.current().panes[0].slim = true

	if err := le.apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}

	loadLayoutConfig()
	spec := layoutSpec(screenTrade)
	if spec.Layout != "grid" {
		t.Fatalf("layout should persist as grid, got %q", spec.Layout)
	}
	if spec.LayoutParams.Int("cols", 0) != 4 {
		t.Fatalf("cols should persist as 4, got %d", spec.LayoutParams.Int("cols", 0))
	}
	if len(spec.Panes) == 0 || !spec.Panes[0].Slim {
		t.Fatalf("first panel should persist as slim: %+v", spec.Panes)
	}
}
