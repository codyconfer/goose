package game

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/notify"
	"github.com/codyconfer/goose/internal/store"
)

func renderForHints(m Model) string {
	m.width = theme.MinScreenWidth
	m.height = 80
	return m.View()
}

func assertHintText(t *testing.T, view string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func TestGameScreenLegendShowsAliases(t *testing.T) {
	s := leveledState()
	s.Tokens = 100000
	m := New(economy.FromState(s), events.NewMachine(), 0)

	assertHintText(t, renderForHints(m),
		"enter/space",
		"↑/↓/j/k",
		"b/→/l",
		"esc/q",
		"O/C",
		"max call",
		"P",
		"max put",
	)
}

func TestTradeDeskLegendShowsAliases(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}

	assertHintText(t, renderForHints(m),
		"←/→/h/l",
		"↑/↓/j/k",
		"enter/space",
		"esc/t/q",
	)
}

func TestSpecDeskLegendShowsAliases(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}

	assertHintText(t, renderForHints(m),
		"←/→/h/l",
		"↑/↓/j/k",
		"[ ]/-/+",
		"enter/space",
		"esc/d/q",
	)
}

func TestAgentsLegendShowsAliases(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.screen = &agentsScreen{prev: &gameScreen{}}

	assertHintText(t, renderForHints(m),
		"↑/↓/j/k",
		"enter/space",
		"←/→/h/l",
		"[ ]/-/+",
		"esc/a/q",
	)
}

func TestCharacterLegendsShowAliases(t *testing.T) {
	base := New(economy.NewMachine(), events.NewMachine(), 0)
	char := &characters.Character{
		Headline: "A Board Member Appears",
		Pitch:    "Choose what to do.",
		Options: []characters.Option{{
			Label: "Take the deal",
			Desc:  "It seems suspicious.",
		}},
	}

	base.screen = &characterScreen{char: char}
	assertHintText(t, renderForHints(base),
		"↑/↓/j/k",
		"enter/space",
	)

	base.screen = &characterScreen{
		char: char,
		notification: &notify.Notification{
			Title:   "Outcome",
			Message: "That happened.",
		},
	}
	assertHintText(t, renderForHints(base),
		"enter/space/esc/q",
	)
}

func TestMenuLegendsShowModeSpecificKeys(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)

	m.screen = &menuScreen{items: menuItems(nil)}
	assertHintText(t, renderForHints(m),
		"↑/↓/j/k",
		"enter/space",
		"x/d",
		"esc/q",
	)

	m.screen = &menuScreen{
		items:  menuItems(nil),
		mode:   menuModeRename,
		target: store.SaveInfo{Name: "Alpha"},
		edit:   "Alpha",
	}
	assertHintText(t, renderForHints(m),
		"type",
		"backspace",
		"enter",
		"esc",
	)

	m.screen = &menuScreen{
		items:  menuItems(nil),
		mode:   menuModeDelete,
		target: store.SaveInfo{Name: "Alpha"},
	}
	assertHintText(t, renderForHints(m),
		"y",
		"n/esc/q",
	)
}

func TestSettingsLegendTracksSeedEditingKeys(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	ss := newSettingsScreen()

	m.screen = ss
	assertHintText(t, renderForHints(m),
		"↑/↓/j/k",
		"←/→/h/l",
		"enter/space",
		"esc/q",
	)

	ss.cursor = len(ss.rows)
	assertHintText(t, renderForHints(m),
		"digits/-",
		"backspace",
		"r",
	)
}
