package game

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func atMinHeight(m Model) string {
	m.width = theme.MinScreenWidth
	m.height = theme.MinBodyHeight
	return m.View()
}

func assertFitsMinHeight(t *testing.T, name, view string) {
	t.Helper()
	if h := lipgloss.Height(view); h > theme.MinBodyHeight {
		t.Errorf("%s rendered %d rows at the minimum height (max %d):\n%s", name, h, theme.MinBodyHeight, view)
	}
}

func assertContains(t *testing.T, name, view string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Errorf("%s missing essential marker %q at the minimum height:\n%s", name, want, view)
		}
	}
}

func TestGameScreenFitsMinHeight(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1_000_000
	s.Eggs = 500
	s.Owned["server"] = 5
	m := New(economy.FromState(s), events.NewMachine(), 0)

	v := atMinHeight(m)
	assertFitsMinHeight(t, "game screen", v)

	assertContains(t, "game screen", v, "🪿", "generate")
}

func TestGameScreenFrozenFitsMinHeight(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1_000_000
	s.Eggs = 500
	s.Owned["server"] = 5
	s.FreezeSeconds = 30
	s.FreezeReason = "a subcommittee is asking questions"
	m := New(economy.FromState(s), events.NewMachine(), 0)

	v := atMinHeight(m)
	assertFitsMinHeight(t, "frozen game screen", v)

	assertContains(t, "frozen game screen", v, "SHUT DOWN", "🪿")
}

func TestTradeDeskFitsMinHeightAndDropsChart(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	econ.ScheduleTrade(economy.TxBuyEggs, 100)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}

	v := atMinHeight(m)
	assertFitsMinHeight(t, "trade desk", v)

	assertContains(t, "trade desk", v, "NEW ORDER", "TRADE QUEUE")
	if strings.Contains(v, "EGG PRICE") {
		t.Errorf("price chart should be dropped at the minimum height:\n%s", v)
	}
}

func TestTradeDeskShowsEverythingWhenTall(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	econ.ScheduleTrade(economy.TxBuyEggs, 100)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	m.width = theme.MinScreenWidth
	m.height = 120

	v := m.View()
	for _, want := range []string{"EGG PRICE", "NEW ORDER", "TRADE QUEUE", "LEDGER"} {
		if !strings.Contains(v, want) {
			t.Errorf("tall trade desk missing %q (should not be dropped):\n%s", want, v)
		}
	}
}

func TestSpecDeskFitsMinHeight(t *testing.T) {
	econ := economy.FromState(leveledState())
	econ.OpenPosition(economy.PosCall, 50, 5, 60)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}

	v := atMinHeight(m)
	assertFitsMinHeight(t, "spec desk", v)

	assertContains(t, "spec desk", v, "WRITE A CONTRACT", "OPEN POSITIONS")
}

func TestGameScreenShortTierDropsMediumPanels(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1_000_000
	s.Eggs = 500
	s.Owned["server"] = 5
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth
	m.height = 20

	v := m.View()
	if h := lipgloss.Height(v); h > m.height {
		t.Errorf("short game screen rendered %d rows, exceeds height %d:\n%s", h, m.height, v)
	}
	assertContains(t, "short game screen", v, "🪿", "generate")
	if strings.Contains(v, content.Text.Market.Panel) {
		t.Errorf("market (tall) should be dropped in short tier:\n%s", v)
	}
}

func TestStartupUnknownHeightFitsEssentials(t *testing.T) {

	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.width = theme.MinScreenWidth

	v := m.View()
	assertFitsMinHeight(t, "startup", v)
	assertContains(t, "startup", v, "🪿", "generate")
}
