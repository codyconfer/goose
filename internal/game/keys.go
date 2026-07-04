package game

import (
	"github.com/codyconfer/viewkit/keys"

	"github.com/codyconfer/goose/internal/content"
)

const (
	actBuy        keys.Action = "game.buy"
	actSell       keys.Action = "game.sell"
	actMaxBuy     keys.Action = "game.max_buy"
	actMaxSell    keys.Action = "game.max_sell"
	actMaxCall    keys.Action = "game.max_call"
	actMaxPut     keys.Action = "game.max_put"
	actOpenTrade  keys.Action = "game.open_trade"
	actOpenAgents keys.Action = "game.open_agents"
	actOpenLayout keys.Action = "game.open_layout"

	actToggleKind  keys.Action = "desk.toggle_kind"
	actCancelOrder keys.Action = "desk.cancel"
	actClearQueue  keys.Action = "trade.clear"
	actOpenSpec    keys.Action = "trade.open_spec"
	actClosePos    keys.Action = "spec.close"
	actCloseAll    keys.Action = "spec.close_all"

	actMenuNew    keys.Action = "menu.new"
	actMenuRename keys.Action = "menu.rename"
	actMenuDelete keys.Action = "menu.delete"
	actMenuSave   keys.Action = "menu.save"
	actMenuLayout keys.Action = "menu.layout"
	actConfirmYes keys.Action = "menu.confirm_delete"

	actReroll keys.Action = "settings.reroll"

	actLayoutSave keys.Action = "layout.save"
)

func toggleBinding(action keys.Action, label string) keys.Binding {
	sc := keys.Cur()
	left, right := sc.Binding(keys.Left), sc.Binding(keys.Right)
	combined := append(append([]string{}, left.Keys...), right.Keys...)
	return keys.Binding{Keys: combined, Action: action, Glyph: left.Glyph, Label: label}
}

func gameKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c", "q", "esc"}, Action: keys.Quit, Glyph: "esc/q", Label: "quit"},
		sc.Binding(keys.Confirm).WithLabel("generate"),
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.FocusNext),
		sc.Binding(keys.FocusPrev),
		keys.Binding{Keys: []string{"b", "right", "l"}, Action: actBuy, Glyph: "b/→/l", Label: "buy"},
		keys.Binding{Keys: []string{"s"}, Action: actSell, Glyph: "s", Label: "sell"},
		keys.Binding{Keys: []string{"B"}, Action: actMaxBuy, Glyph: "B/S", Label: "max queue"},
		keys.Binding{Keys: []string{"S"}, Action: actMaxSell},
		keys.Binding{Keys: []string{"O", "C"}, Action: actMaxCall, Glyph: "O/C", Label: "max call"},
		keys.Binding{Keys: []string{"P"}, Action: actMaxPut, Glyph: "P", Label: "max put"},
		keys.Binding{Keys: []string{"t"}, Action: actOpenTrade, Glyph: "t", Label: "trade"},
		keys.Binding{Keys: []string{"a"}, Action: actOpenAgents, Glyph: "a", Label: "agents"},
		keys.Binding{Keys: []string{"L"}, Action: actOpenLayout, Glyph: "L", Label: "layout"},
	)
}

func characterKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		sc.Binding(keys.Up).WithLabel("weigh options"),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm).WithLabel("decide"),
	)
}

func characterNotifyKeymap() *keys.Map {
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{
			Keys:   []string{"enter", " ", "spacebar", "esc", "q"},
			Action: keys.Confirm,
			Glyph:  "enter/space/esc/q",
			Label:  content.Text.Character.BackHint,
		},
	)
}

func agentsKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{Keys: []string{"esc", "a", "q"}, Action: keys.Cancel, Glyph: "esc/a/q", Label: "back"},
		sc.Binding(keys.Up).WithLabel("select"),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm).WithLabel("hire/bench"),
		sc.Binding(keys.Left).WithLabel("size"),
		sc.Binding(keys.Right),
		sc.Binding(keys.Inc).WithLabel("threshold"),
		sc.Binding(keys.Dec),
	)
}

func tradeKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{Keys: []string{"esc", "t", "q"}, Action: keys.Cancel, Glyph: "esc/t/q", Label: "back"},
		toggleBinding(actToggleKind, "buy/sell"),
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.FocusNext),
		sc.Binding(keys.FocusPrev),
		sc.Binding(keys.Confirm).WithLabel("queue"),
		keys.Binding{Keys: []string{"x"}, Action: actCancelOrder, Glyph: "x", Label: "cancel"},
		keys.Binding{Keys: []string{"c"}, Action: actClearQueue, Glyph: "c", Label: "clear"},
		keys.Binding{Keys: []string{"d"}, Action: actOpenSpec, Glyph: "d", Label: "derivatives"},
	)
}

func specKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{Keys: []string{"esc", "d", "q"}, Action: keys.Cancel, Glyph: "esc/d/q", Label: "back"},
		toggleBinding(actToggleKind, "call/put"),
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.FocusNext),
		sc.Binding(keys.FocusPrev),
		sc.Binding(keys.Inc).WithLabel("leverage"),
		sc.Binding(keys.Dec),
		sc.Binding(keys.Confirm).WithLabel("open"),
		keys.Binding{Keys: []string{"x"}, Action: actClosePos, Glyph: "x", Label: "close"},
		keys.Binding{Keys: []string{"c"}, Action: actCloseAll, Glyph: "c", Label: "close all"},
	)
}

func menuKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c", "q", "esc"}, Action: keys.Quit, Glyph: "esc/q", Label: "quit"},
		keys.Binding{Keys: []string{"n"}, Action: actMenuNew, Glyph: "n", Label: "new"},
		keys.Binding{Keys: []string{"r"}, Action: actMenuRename, Glyph: "r", Label: "rename"},
		keys.Binding{Keys: []string{"x", "d"}, Action: actMenuDelete, Glyph: "x/d", Label: "delete"},
		keys.Binding{Keys: []string{"l"}, Action: actMenuLayout, Glyph: "l", Label: "layout"},
		sc.Binding(keys.Up).WithLabel("select"),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm).WithLabel("choose"),
	)
}

func layoutEditorKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{Keys: []string{"esc", "q"}, Action: keys.Cancel, Glyph: "esc/q", Label: "back"},
		sc.Binding(keys.Up).WithLabel("row"),
		sc.Binding(keys.Down),
		sc.Binding(keys.Left).WithLabel("change"),
		sc.Binding(keys.Right),
		sc.Binding(keys.Confirm).WithLabel("toggle panel"),
		sc.Binding(keys.Inc).WithLabel("reorder"),
		sc.Binding(keys.Dec),
		keys.Binding{Keys: []string{"w"}, Action: actLayoutSave, Glyph: "w", Label: "save"},
	)
}

func menuRenameKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"enter"}, Action: actMenuSave, Glyph: "enter", Label: "save"},
		keys.Binding{Keys: []string{"esc", "ctrl+c"}, Action: keys.Cancel, Glyph: "esc", Label: "cancel"},
		sc.Binding(keys.Erase).WithLabel("erase"),
	)
}

func menuDeleteKeymap() *keys.Map {
	return keys.NewMap(
		keys.Binding{Keys: []string{"y", "Y"}, Action: actConfirmYes, Glyph: "y", Label: "delete"},
		keys.Binding{Keys: []string{"n", "N", "esc", "q", "ctrl+c"}, Action: keys.Cancel, Glyph: "n/esc/q", Label: "cancel"},
	)
}

func settingsKeymap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		keys.Binding{Keys: []string{"ctrl+c"}, Action: keys.Quit},
		keys.Binding{Keys: []string{"esc", "q"}, Action: keys.Cancel, Glyph: "esc/q", Label: "back"},
		sc.Binding(keys.Up).WithLabel("setting"),
		sc.Binding(keys.Down),
		sc.Binding(keys.Left).WithLabel("change"),
		sc.Binding(keys.Right),
		sc.Binding(keys.Confirm).WithLabel("hatch flock"),
		sc.Binding(keys.Erase).WithLabel("erase"),
		keys.Binding{Keys: []string{"r"}, Action: actReroll, Glyph: "r", Label: "reroll seed"},
	)
}
