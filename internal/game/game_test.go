package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/notify"
	"github.com/codyconfer/goose/internal/store"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func send(m Model, k tea.KeyMsg) Model {
	next, _ := m.Update(k)
	return next.(Model)
}

func maxLineWidth(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		if w := ansi.StringWidth(line); w > max {
			max = w
		}
	}
	return max
}

func TestTapEarnsTokens(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	before := m.econ.Get().Tokens
	m = send(m, key("enter"))
	if m.econ.Get().Tokens != before+m.econ.Get().PerClick {
		t.Fatalf("tap earned %v, want +%v", m.econ.Get().Tokens-before, m.econ.Get().PerClick)
	}
	if m.pulse <= 0 {
		t.Fatal("tap did not trigger the pulse animation")
	}
}

func TestCapexNavigationSkipsLocked(t *testing.T) {

	m := New(economy.NewMachine(), events.NewMachine(), 0)
	if c := m.screen.(*gameScreen).cursor; c != 0 {
		t.Fatalf("start cursor=%d, want 0", c)
	}
	m = send(m, key("down"))
	if c := m.screen.(*gameScreen).cursor; c != 3 {
		t.Fatalf("after down cursor=%d, want 3 (skipped locked crier + blitz)", c)
	}
}

func TestBuyFromCapex(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 100
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("b"))
	if m.econ.Get().UpgradeLevel(economy.UpgradeClick) != 1 {
		t.Fatalf("click level=%d, want 1", m.econ.Get().UpgradeLevel(economy.UpgradeClick))
	}
	if m.econ.Get().PerClick != 2 {
		t.Fatalf("per-click after buy=%v, want 2", m.econ.Get().PerClick)
	}
	if m.flash == "" {
		t.Fatal("expected a purchase flash message")
	}
}

func TestHeartbeatSimulatesOnlyInGame(t *testing.T) {

	s := economy.NewState()
	s.Owned["server"] = 5
	g := New(economy.FromState(s), events.NewMachine(), 0)
	g = func() Model { n, _ := g.Update(upBeatMsg(time.Now().Add(time.Second))); return n.(Model) }()
	if g.econ.Get().Tokens <= 0 {
		t.Fatalf("game heartbeat did not earn tokens: %v", g.econ.Get().Tokens)
	}

	isolateHome(t)
	mm := NewMenu()
	ms := economy.NewState()
	ms.Owned["server"] = 5
	mm.econ = economy.FromState(ms)
	before := mm.econ.Get().Tokens
	mm = func() Model { n, _ := mm.Update(upBeatMsg(time.Now().Add(time.Second))); return n.(Model) }()
	if mm.econ.Get().Tokens != before {
		t.Fatalf("menu heartbeat changed tokens: %v -> %v", before, mm.econ.Get().Tokens)
	}
}

func TestPriceReRollsOnlyOnSlowBeat(t *testing.T) {
	s := economy.NewState()
	s.Owned["server"] = 5
	s.Consumers = 50
	m := New(economy.FromState(s), events.NewMachine(), 0)

	base := time.Unix(1_700_000_000, 0)
	m.rng = rand.New(rand.NewSource(1))
	m.clock = newClock(base)

	for i := 1; i <= 9; i++ {
		n, _ := m.Update(upBeatMsg(base.Add(time.Duration(i) * upBeatRate)))
		m = n.(Model)
	}
	if m.econ.Get().PriceFactor != 1 {
		t.Fatalf("price moved before a slow beat: %v", m.econ.Get().PriceFactor)
	}

	n, _ := m.Update(upBeatMsg(base.Add(10 * upBeatRate)))
	m = n.(Model)
	if m.econ.Get().PriceFactor == 1 {
		t.Fatal("price did not re-roll on the slow beat")
	}
}

func TestMenuStartNewGame(t *testing.T) {
	isolateHome(t)
	m := NewMenu()
	m = send(m, key("enter"))
	if _, ok := m.screen.(*settingsScreen); !ok {
		t.Fatalf("after New, screen=%T, want *settingsScreen", m.screen)
	}
	m = send(m, key("enter"))
	if _, ok := m.screen.(*gameScreen); !ok {
		t.Fatalf("after confirming settings, screen=%T, want *gameScreen", m.screen)
	}
	if m.saveID <= 0 || m.saveName == "" {
		t.Fatalf("new game did not create a save slot: id=%d name=%q", m.saveID, m.saveName)
	}
}

func TestSettingsScreenAppliesPacing(t *testing.T) {
	isolateHome(t)
	m := NewMenu()
	m = send(m, key("enter"))
	ss, ok := m.screen.(*settingsScreen)
	if !ok {
		t.Fatalf("expected settings screen, got %T", m.screen)
	}

	m = send(m, key("down"))
	m = send(m, key("down"))
	want := ss.rows[2].idx + 1
	m = send(m, key("right"))
	m = send(m, key("enter"))
	if got := m.econ.Get().Settings.MarketIdx(); got != want {
		t.Fatalf("market pace idx=%d, want %d", got, want)
	}
}

func TestEmptySettingsScreenUsesDefaults(t *testing.T) {
	isolateHome(t)
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.screen = &settingsScreen{}
	_ = m.View()
	m = send(m, key("right"))
	m = send(m, key("enter"))
	if _, ok := m.screen.(*gameScreen); !ok {
		t.Fatalf("empty settings screen opened %T, want *gameScreen", m.screen)
	}
}

func TestMenuExitQuits(t *testing.T) {
	isolateHome(t)
	m := NewMenu()
	items := len(m.screen.(*menuScreen).items)
	for i := 0; i < items; i++ {
		m = send(m, key("down"))
	}
	m = send(m, key("enter"))
	if !m.quitting {
		t.Fatal("selecting Exit did not set quitting")
	}
}

func TestMenuCursorClamps(t *testing.T) {
	isolateHome(t)
	m := NewMenu()
	m = send(m, key("up"))
	if c := m.screen.(*menuScreen).cursor; c != 0 {
		t.Fatalf("cursor went above top: %d", c)
	}
}

func TestMenuManagesNamedSaves(t *testing.T) {
	isolateHome(t)
	state := economy.NewState()
	state.Tokens = 99
	info, err := store.CreateSave("Alpha", economy.FromState(state), events.NewMachine())
	if err != nil {
		t.Fatalf("create save: %v", err)
	}

	m := NewMenu()
	ms := m.screen.(*menuScreen)
	if len(ms.items) < 1 || ms.items[0].action != menuSave {
		t.Fatalf("first item=%+v, want save", ms.items)
	}
	m = send(m, key("enter"))
	if m.saveID != info.ID || m.saveName != "Alpha" || m.econ.Get().Tokens != 99 {
		t.Fatalf("loaded save mismatch: id=%d name=%q tokens=%v", m.saveID, m.saveName, m.econ.Get().Tokens)
	}

	m = NewMenu()
	m = send(m, key("r"))
	for range "Alpha" {
		m = send(m, key("backspace"))
	}
	m = send(m, key("Beta"))
	m = send(m, key("enter"))
	saves, err := store.ListSaves()
	if err != nil {
		t.Fatalf("list saves: %v", err)
	}
	if len(saves) != 1 || saves[0].Name != "Beta" {
		t.Fatalf("renamed saves=%+v", saves)
	}

	m = NewMenu()
	m = send(m, key("x"))
	m = send(m, key("y"))
	saves, err = store.ListSaves()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(saves) != 0 {
		t.Fatalf("delete left saves=%+v", saves)
	}
}

func TestVCFlowResolveAndDismiss(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 5000
	s.TotalEarned = 5000
	m := New(economy.FromState(s), events.NewMachine(), 0)
	ch := characters.NewVC(m.econ.Get(), m.rng)
	prev := &gameScreen{cursor: 2}
	m.screen = &characterScreen{char: &ch, prev: prev}

	m = send(m, key("down"))
	m = send(m, key("enter"))
	cs, ok := m.screen.(*characterScreen)
	if !ok || cs.notification == nil {
		t.Fatalf("expected a resolved characterScreen, got %T", m.screen)
	}

	m = send(m, key("enter"))
	if m.screen != prev {
		t.Fatalf("did not return to the prior game screen: %T", m.screen)
	}
	if !m.notifs.Active() {
		t.Fatal("VC outcome did not leave a notification banner")
	}
}

func TestViewRendersEveryScreen(t *testing.T) {
	isolateHome(t)
	if !strings.Contains(NewMenu().View(), "GOLDEN GOOSE") {
		t.Error("menu view missing title")
	}

	g := New(economy.NewMachine(), events.NewMachine(), 0)
	gv := g.View()
	for _, want := range []string{"FLOCK", "CAPEX", "Enter the Flow State"} {
		if !strings.Contains(gv, want) {
			t.Errorf("game view missing %q", want)
		}
	}

	s := economy.NewState()
	s.Tokens = 5000
	ch := characters.NewVC(s, g.rng)
	g.screen = &characterScreen{char: &ch, prev: &gameScreen{}}
	if !strings.Contains(g.View(), "VENTURE CAPITALIST") {
		t.Error("character view missing header")
	}

	g.quitting = true
	if !strings.Contains(g.View(), "powers down") {
		t.Error("quitting view missing farewell")
	}
}

func TestNarrowMenuViewBoundsLongSaveNames(t *testing.T) {
	isolateHome(t)
	name := strings.Repeat("LongSaveName", 8)
	state := economy.NewState()
	state.Tokens = 42
	if _, err := store.CreateSave(name, economy.FromState(state), events.NewMachine()); err != nil {
		t.Fatalf("create save: %v", err)
	}

	m := NewMenu()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 48, Height: 20})
	m = next.(Model)
	view := m.View()
	if strings.Contains(view, name) {
		t.Fatalf("view rendered unbounded save name")
	}
	if got := maxLineWidth(view); got > 52 {
		t.Fatalf("max line width=%d, want <= 52:\n%s", got, view)
	}
}

func TestInitStartsHeartbeat(t *testing.T) {
	if New(economy.NewMachine(), events.NewMachine(), 0).Init() == nil {
		t.Fatal("Init should return the heartbeat command")
	}
}

func TestActiveGameRenderAndProducerBuy(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1_000_000
	s.Eggs = 500
	s.Owned["server"] = 5
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m.notifs.Push(notify.Notification{Title: "🍀 Test Notification", Message: "hi", Tone: notify.TonePositive}, 5)

	v := m.View()
	if !strings.Contains(v, "EGG MARKET") {
		t.Error("market panel not rendered when eggs are present")
	}
	if !strings.Contains(v, "Test Notification") {
		t.Error("notification banner not rendered")
	}

	m = send(m, key("down"))
	m = send(m, key("down"))
	_ = m.View()
	m = send(m, key("up"))
	m = send(m, key("down"))
	m = send(m, key("down"))
	m = send(m, key("b"))
	if m.econ.Get().Count("gpu") < 1 {
		t.Fatalf("expected to buy a GPU, owned=%d", m.econ.Get().Count("gpu"))
	}
}

func TestBuyDeniedWhenBroke(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m = send(m, key("b"))
	if !strings.Contains(strings.ToLower(m.flash), "not enough") {
		t.Fatalf("upgrade denial flash=%q", m.flash)
	}
	m = send(m, key("down"))
	m = send(m, key("b"))
	if !strings.Contains(strings.ToLower(m.flash), "not enough") {
		t.Fatalf("producer denial flash=%q", m.flash)
	}
}

func TestBeatRollsAndSpawnsVC(t *testing.T) {
	s := economy.NewState()
	s.TotalEarned = 5000
	s.Owned["server"] = 3
	s.Tokens = 1000
	m := New(economy.FromState(s), events.NewMachine(), 0)

	m.beatMid()

	for i := 0; i < 5000; i++ {
		if _, ok := m.screen.(*characterScreen); ok {
			break
		}
		m.beatChars()
	}
	if _, ok := m.screen.(*characterScreen); !ok {
		t.Fatal("beatChars never spawned a character over many rolls")
	}
}

func TestScreenSimulates(t *testing.T) {
	if (&menuScreen{}).simulates() {
		t.Error("menu should not simulate")
	}
	if !(&gameScreen{}).simulates() {
		t.Error("game should simulate")
	}
	if !(&characterScreen{}).simulates() {
		t.Error("character screen should simulate")
	}
}
