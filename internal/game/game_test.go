package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/notify"
	"github.com/codyconfer/goose/internal/store"
	"github.com/codyconfer/goose/internal/world"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
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
	for line := range strings.SplitSeq(s, "\n") {
		if w := ansi.StringWidth(line); w > max {
			max = w
		}
	}
	return max
}

func mustBuildCharacter(t *testing.T, m Model, key string, s economy.State) characters.Character {
	t.Helper()
	for _, spec := range m.world.Characters {
		if spec.Key == key {
			return spec.Build(s)
		}
	}
	t.Fatalf("missing character %q", key)
	return characters.Character{}
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
	if !m.feed.active() {
		t.Fatal("expected a purchase message in the ticker feed")
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

func TestAgentFireLandsInFeedNotFlash(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	s.Agents = []economy.Agent{
		{Key: "momentum", Enabled: true, Metric: economy.MetricTokens, Cmp: economy.CmpAbove, Threshold: 0, Action: economy.ActBuyEggs, Size: 10},
	}
	m := New(economy.FromState(s), events.NewMachine(), 0)
	base := time.Unix(1_700_000_000, 0)
	m.rng = rand.New(rand.NewSource(1))
	m.clock = newClock(base)

	for i := 1; i <= 10; i++ {
		n, _ := m.Update(upBeatMsg(base.Add(time.Duration(i) * upBeatRate)))
		m = n.(Model)
	}

	if !m.feed.active() {
		t.Fatal("agent fire did not reach the feed")
	}
	if m.flash != "" {
		t.Fatalf("agent fire leaked into flash: %q", m.flash)
	}
	if !strings.Contains(m.renderFeed(0, false), content.Text.Feed.Panel) {
		t.Fatalf("feed panel not rendered: %q", m.renderFeed(0, false))
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

func TestBaselineYieldAccruesOnlyOnSlowBeat(t *testing.T) {

	m := New(economy.NewMachine(), events.NewMachine(), 0)

	base := time.Unix(1_700_000_000, 0)
	m.rng = rand.New(rand.NewSource(1))
	m.clock = newClock(base)

	for i := 1; i <= 9; i++ {
		n, _ := m.Update(upBeatMsg(base.Add(time.Duration(i) * upBeatRate)))
		m = n.(Model)
	}
	if s := m.econ.Get(); s.Eggs != 0 || s.Tokens != 0 {
		t.Fatalf("baseline accrued before slow beat: eggs=%v tokens=%v", s.Eggs, s.Tokens)
	}

	n, _ := m.Update(upBeatMsg(base.Add(10 * upBeatRate)))
	m = n.(Model)
	if s := m.econ.Get(); s.Eggs != 1 || s.Tokens != 1 {
		t.Fatalf("baseline did not drip on slow beat: eggs=%v tokens=%v", s.Eggs, s.Tokens)
	}
}

func TestOfflineNoteExpiresAfterTTL(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 500)
	if m.offline != 500 || m.offlineTTL != offlineBeats {
		t.Fatalf("offline not armed on load: offline=%v ttl=%d", m.offline, m.offlineTTL)
	}

	base := time.Unix(1_700_000_000, 0)
	m.rng = rand.New(rand.NewSource(1))
	m.clock = newClock(base)

	for i := 1; i < offlineBeats; i++ {
		n, _ := m.Update(upBeatMsg(base.Add(time.Duration(i) * upBeatRate)))
		m = n.(Model)
	}
	if m.offline == 0 {
		t.Fatalf("offline note cleared before TTL elapsed (beat %d/%d)", offlineBeats-1, offlineBeats)
	}

	n, _ := m.Update(upBeatMsg(base.Add(time.Duration(offlineBeats) * upBeatRate)))
	m = n.(Model)
	if m.offline != 0 || m.offlineTTL != 0 {
		t.Fatalf("offline note did not expire at TTL: offline=%v ttl=%d", m.offline, m.offlineTTL)
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

func TestMenuDeletesSaveCreatedThroughNewGameFlow(t *testing.T) {
	isolateHome(t)

	m := NewMenu()
	m = send(m, key("enter"))
	if _, ok := m.screen.(*settingsScreen); !ok {
		t.Fatalf("after New, screen=%T, want *settingsScreen", m.screen)
	}
	m = send(m, key("enter"))
	if m.saveID <= 0 {
		t.Fatalf("new game did not create a save: id=%d", m.saveID)
	}

	m = NewMenu()
	ms, ok := m.screen.(*menuScreen)
	if !ok {
		t.Fatalf("menu screen=%T, want *menuScreen", m.screen)
	}
	if len(ms.items) == 0 || ms.items[0].action != menuSave {
		t.Fatalf("menu items=%+v, want first item to be a save", ms.items)
	}

	m = send(m, key("x"))
	ms = m.screen.(*menuScreen)
	if ms.mode != menuModeDelete {
		t.Fatalf("menu mode=%v, want delete confirmation", ms.mode)
	}
	m = send(m, key("y"))

	saves, err := store.ListSaves()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(saves) != 0 {
		t.Fatalf("delete left saves=%+v", saves)
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
	for range items {
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
	info, err := store.CreateSave("Alpha", economy.FromState(state), events.NewMachine(), world.Generate(world.DefaultSeed))
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
	ch := mustBuildCharacter(t, m, "vc", m.econ.Get())
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
	ch := mustBuildCharacter(t, g, "vc", s)
	g.screen = &characterScreen{char: &ch, prev: &gameScreen{}}
	if !strings.Contains(g.View(), "VENTURE CAPITALIST") {
		t.Error("character view missing header")
	}

	g.screen = &agentsScreen{prev: &gameScreen{}}
	if !strings.Contains(g.View(), "AGENT DESK") {
		t.Error("agents view missing header")
	}

	g.quitting = true
	if !strings.Contains(g.View(), "powers down") {
		t.Error("quitting view missing farewell")
	}
}

func TestNarrowMenuViewRequestsWiderTerminal(t *testing.T) {
	isolateHome(t)
	name := strings.Repeat("LongSaveName", 8)
	state := economy.NewState()
	state.Tokens = 42
	if _, err := store.CreateSave(name, economy.FromState(state), events.NewMachine(), world.Generate(world.DefaultSeed)); err != nil {
		t.Fatalf("create save: %v", err)
	}

	m := NewMenu()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 48, Height: 20})
	m = next.(Model)
	view := m.View()
	if strings.Contains(view, name) {
		t.Fatalf("view rendered save content despite a too-narrow screen")
	}
	if !strings.Contains(view, "TERMINAL TOO NARROW") {
		t.Fatalf("view did not request a wider terminal:\n%s", view)
	}
	if got := maxLineWidth(view); got > 48 {
		t.Fatalf("max line width=%d, want <= 48:\n%s", got, view)
	}
}

func TestViewRequiresMinimumScreenWidthForExistingScreens(t *testing.T) {
	isolateHome(t)

	s := economy.NewState()
	s.Tokens = 5000
	seededGame := New(economy.FromState(s), events.NewMachine(), 0)
	ch := mustBuildCharacter(t, seededGame, "vc", s)

	tests := []struct {
		name   string
		build  func() Model
		hidden string
	}{
		{
			name: "menu",
			build: func() Model {
				return NewMenu()
			},
			hidden: "GOLDEN GOOSE",
		},
		{
			name: "game",
			build: func() Model {
				return New(economy.NewMachine(), events.NewMachine(), 0)
			},
			hidden: "CAPEX",
		},
		{
			name: "settings",
			build: func() Model {
				m := NewMenu()
				m.screen = newSettingsScreen()
				return m
			},
			hidden: "NEW FLOCK",
		},
		{
			name: "trade",
			build: func() Model {
				m := New(economy.NewMachine(), events.NewMachine(), 0)
				m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
				return m
			},
			hidden: "TRADE DESK",
		},
		{
			name: "spec",
			build: func() Model {
				m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
				m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
				return m
			},
			hidden: "DERIVATIVES DESK",
		},
		{
			name: "character",
			build: func() Model {
				m := New(economy.FromState(s), events.NewMachine(), 0)
				m.screen = &characterScreen{char: &ch, prev: &gameScreen{}}
				return m
			},
			hidden: "VENTURE CAPITALIST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.build()
			next, _ := m.Update(tea.WindowSizeMsg{Width: theme.MinScreenWidth - 1, Height: 24})
			m = next.(Model)

			view := m.View()
			for _, want := range []string{"TERMINAL TOO NARROW", "80", "79"} {
				if !strings.Contains(view, want) {
					t.Fatalf("view missing %q:\n%s", want, view)
				}
			}
			if strings.Contains(view, tt.hidden) {
				t.Fatalf("view rendered %q on a too-narrow screen:\n%s", tt.hidden, view)
			}
		})
	}
}

func TestPageScrollRevealsOffscreenPanels(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	for i := 0; i < 12; i++ {
		econ.ScheduleTrade(economy.TxBuyEggs, float64(i+1)*10)
	}
	econ.BuyProducer(economy.Producers[0])

	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	next, _ := m.Update(tea.WindowSizeMsg{Width: theme.MinScreenWidth, Height: 18})
	m = next.(Model)

	if view := m.View(); strings.Contains(view, "TRADE QUEUE") {
		t.Fatalf("queue should start below the viewport:\n%s", view)
	}

	if view := m.View(); !strings.Contains(view, "pgup/pgdn") {
		t.Fatalf("viewport scroll hint missing before scroll:\n%s", view)
	}
	if view := m.View(); !strings.Contains(view, "esc/t/q") {
		t.Fatalf("sticky footer legend missing before scroll:\n%s", view)
	}

	view := m.View()
	for range 10 {
		if strings.Contains(view, "TRADE QUEUE") {
			break
		}
		m = send(m, key("pgdown"))
		view = m.View()
	}
	if !strings.Contains(view, "TRADE QUEUE") || !strings.Contains(view, "esc/t/q") {
		t.Fatalf("page scroll did not keep footer visible while revealing lower panels:\n%s", view)
	}
}

func TestGameScreenCapexPanelScrollsToSelection(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1_000_000_000
	s.PeakEggs = economy.LevelThresholds[len(economy.LevelThresholds)-1]

	m := New(economy.FromState(s), events.NewMachine(), 0)
	for range 12 {
		m = send(m, key("down"))
	}

	gs := m.screen.(*gameScreen)
	if gs.capex.Offset == 0 {
		_ = m.View()
	}
	if gs.capex.Offset == 0 {
		t.Fatal("capex selection did not advance the panel scroll")
	}
	if got := m.View(); !strings.Contains(got, "↕") {
		t.Fatalf("capex panel missing scroll footer:\n%s", got)
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
	m.height = 120
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
	if feed := strings.ToLower(strings.Join(m.feed.lines(), "\n")); !strings.Contains(feed, "not enough") {
		t.Fatalf("upgrade denial feed=%q", feed)
	}
	m = send(m, key("down"))
	m = send(m, key("b"))
	if feed := strings.ToLower(strings.Join(m.feed.lines(), "\n")); !strings.Contains(feed, "not enough") {
		t.Fatalf("producer denial feed=%q", feed)
	}
}

func TestBeatRollsAndSpawnsVC(t *testing.T) {
	s := economy.NewState()
	s.TotalEarned = 5000
	s.Owned["server"] = 3
	s.Tokens = 1000
	m := New(economy.FromState(s), events.NewMachine(), 0)

	m.beatMid()

	for range 5000 {
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
