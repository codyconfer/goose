package game

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"
)

type editPane struct {
	key   string
	title string
	on    bool
	slim  bool
}

type editSpec struct {
	layoutIdx int
	layouts   []string
	params    map[string]int
	panes     []editPane
}

type paramSpec struct {
	key   string
	label string
	min   int
	max   int
	def   int
}

var layoutTitles = map[string]string{
	"single":       "Single Column",
	"flex-columns": "Flex Columns",
	"flex-rows":    "Flex Rows",
	"grid":         "Grid",
}

func layoutDisplayName(key string) string {
	if t, ok := layoutTitles[key]; ok {
		return t
	}
	return key
}

func layoutParamSpecs(layoutKey string) []paramSpec {
	switch layoutKey {
	case "grid":
		return []paramSpec{
			{key: "cols", label: "Columns", min: 1, max: 4, def: 2},
			{key: "rows", label: "Rows", min: 0, max: 4, def: 0},
		}
	case "flex-columns", "flex-rows":
		return []paramSpec{
			{key: "minWidth", label: "Min Track Width", min: 20, max: 80, def: layout.DefaultFlexMinWidth},
			{key: "maxCols", label: "Max Tracks", min: 1, max: 4, def: layout.DefaultFlexMaxCols},
		}
	}
	return nil
}

type layoutEditorScreen struct {
	prev      screen
	screenIdx int
	cursor    int
	specs     map[string]*editSpec
	themeKeys []string
	themeIdx  int
	clockIdx  int
}

func newLayoutEditor(prev screen) *layoutEditorScreen {
	specs := make(map[string]*editSpec, len(configurableScreens))
	for _, id := range configurableScreens {
		specs[id] = newEditSpec(id)
	}
	tk := theme.Keys()
	return &layoutEditorScreen{prev: prev, specs: specs, themeKeys: tk, themeIdx: currentThemeIndex(tk), clockIdx: currentClockIndex()}
}

func newEditSpec(id string) *editSpec {
	spec := layoutSpec(id)
	infos := paneRegistryKeys(id)
	title := make(map[string]string, len(infos))
	for _, in := range infos {
		title[in.Key] = in.Title
	}

	slim := map[string]bool{}
	seen := map[string]bool{}
	var panes []editPane
	for _, ref := range spec.Panes {
		t, ok := title[ref.Key]
		if !ok {
			continue
		}
		panes = append(panes, editPane{key: ref.Key, title: t, on: true, slim: ref.Slim})
		seen[ref.Key] = true
		slim[ref.Key] = ref.Slim
	}
	for _, in := range infos {
		if seen[in.Key] {
			continue
		}
		panes = append(panes, editPane{key: in.Key, title: in.Title, on: false})
	}

	layouts := layoutKeys(id)
	params := map[string]int{}
	for _, ps := range layoutParamSpecs(spec.Layout) {
		params[ps.key] = spec.LayoutParams.Int(ps.key, ps.def)
	}
	return &editSpec{layoutIdx: layoutIndex(layouts, spec.Layout), layouts: layouts, params: params, panes: panes}
}

func layoutIndex(layouts []string, want string) int {
	for i, l := range layouts {
		if l == want {
			return i
		}
	}
	return 0
}

func (le *layoutEditorScreen) simulates() bool { return false }

func (le *layoutEditorScreen) current() *editSpec {
	return le.specs[configurableScreens[le.screenIdx]]
}

func (le *layoutEditorScreen) currentLayoutKey() string {
	es := le.current()
	if es.layoutIdx >= 0 && es.layoutIdx < len(es.layouts) {
		return es.layouts[es.layoutIdx]
	}
	return "single"
}

func (le *layoutEditorScreen) paramSpecs() []paramSpec {
	return layoutParamSpecs(le.currentLayoutKey())
}

func (le *layoutEditorScreen) themePos() int {
	return 2 + len(le.paramSpecs())
}

func (le *layoutEditorScreen) clockPos() int {
	return le.themePos() + 1
}

func (le *layoutEditorScreen) paneBase() int {
	return le.clockPos() + 1
}

func (le *layoutEditorScreen) stepTheme(delta int) {
	if len(le.themeKeys) == 0 {
		return
	}
	le.themeIdx = panels.StepIndex(le.themeIdx, delta, len(le.themeKeys))
	setTheme(le.themeKeys[le.themeIdx])
	_ = saveThemeConfig()
}

func (le *layoutEditorScreen) stepClock(delta int) {
	if len(clockZoneChoices) == 0 {
		return
	}
	le.clockIdx = panels.StepIndex(le.clockIdx, delta, len(clockZoneChoices))
	setClockZone(le.clockIdx)
}

func (le *layoutEditorScreen) rowCount() int {
	return le.paneBase() + len(le.current().panes)
}

func (le *layoutEditorScreen) paneIndex() (int, bool) {
	p := le.cursor - le.paneBase()
	if p >= 0 && p < len(le.current().panes) {
		return p, true
	}
	return 0, false
}

func (le *layoutEditorScreen) paramValue(ps paramSpec) int {
	if v, ok := le.current().params[ps.key]; ok {
		return v
	}
	return ps.def
}

func (le *layoutEditorScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := layoutEditorKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	switch action {
	case keys.Quit:
		m.quitting = true
		return tea.Quit
	case keys.Cancel:
		m.screen = le.prev
	case keys.Up:
		le.cursor = panels.MoveIndex(le.cursor, -1, le.rowCount())
	case keys.Down:
		le.cursor = panels.MoveIndex(le.cursor, 1, le.rowCount())
	case keys.Left:
		le.changeRow(-1)
	case keys.Right:
		le.changeRow(1)
	case keys.Confirm:
		if p, ok := le.paneIndex(); ok {
			le.current().panes[p].on = !le.current().panes[p].on
		}
	case actLayoutSlim:
		if p, ok := le.paneIndex(); ok {
			le.current().panes[p].slim = !le.current().panes[p].slim
		}
	case keys.Inc:
		le.reorder(1)
	case keys.Dec:
		le.reorder(-1)
	case actLayoutSave:
		if err := le.apply(); err != nil {
			m.setFlash("Couldn't save the layout.")
		} else {
			m.setFlash("Layout saved.")
		}
	}
	return nil
}

func (le *layoutEditorScreen) changeRow(delta int) {
	switch {
	case le.cursor == 0:
		le.screenIdx = panels.StepIndex(le.screenIdx, delta, len(configurableScreens))
		le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	case le.cursor == 1:
		es := le.current()
		if len(es.layouts) > 0 {
			es.layoutIdx = panels.StepIndex(es.layoutIdx, delta, len(es.layouts))
		}
		le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	case le.cursor == le.themePos():
		le.stepTheme(delta)
	case le.cursor == le.clockPos():
		le.stepClock(delta)
	default:
		specs := le.paramSpecs()
		if pi := le.cursor - 2; pi >= 0 && pi < len(specs) {
			ps := specs[pi]
			v := le.paramValue(ps) + delta
			if v < ps.min {
				v = ps.min
			}
			if v > ps.max {
				v = ps.max
			}
			le.current().params[ps.key] = v
		}
	}
}

func (le *layoutEditorScreen) reorder(delta int) {
	p, ok := le.paneIndex()
	if !ok {
		return
	}
	q := p + delta
	es := le.current()
	if q < 0 || q >= len(es.panes) {
		return
	}
	es.panes[p], es.panes[q] = es.panes[q], es.panes[p]
	le.cursor += delta
}

func (le *layoutEditorScreen) apply() error {
	for id, es := range le.specs {
		var panes []layout.PaneRef
		for _, p := range es.panes {
			if p.on {
				panes = append(panes, layout.PaneRef{Key: p.key, Slim: p.slim})
			}
		}
		lay := "single"
		if es.layoutIdx >= 0 && es.layoutIdx < len(es.layouts) {
			lay = es.layouts[es.layoutIdx]
		}
		var lp layout.Params
		if specs := layoutParamSpecs(lay); len(specs) > 0 {
			lp = layout.Params{}
			for _, ps := range specs {
				v := ps.def
				if got, ok := es.params[ps.key]; ok {
					v = got
				}
				lp[ps.key] = v
			}
		}
		setLayoutSpec(id, layout.ScreenSpec{Layout: lay, LayoutParams: lp, Panes: panes})
	}
	return saveLayoutConfig()
}

func (le *layoutEditorScreen) view(m *Model) string {
	vk := m.frame()
	le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	id := configurableScreens[le.screenIdx]
	es := le.current()

	var b strings.Builder
	b.WriteString(vk.Header("SCREEN LAYOUT", "pick a layout, then toggle, reorder, and slim panels"))
	b.WriteString("\n\n")

	b.WriteString(le.selectorRow(vk, "Screen", screenTitles[id], le.cursor == 0, le.screenIdx > 0, le.screenIdx < len(configurableScreens)-1))
	b.WriteString("\n")
	b.WriteString(le.selectorRow(vk, "Layout", layoutDisplayName(le.currentLayoutKey()), le.cursor == 1, es.layoutIdx > 0, es.layoutIdx < len(es.layouts)-1))
	b.WriteString("\n")

	specs := le.paramSpecs()
	for i, ps := range specs {
		v := le.paramValue(ps)
		b.WriteString(le.selectorRow(vk, ps.label, paramDisplay(ps, v), le.cursor == 2+i, v > ps.min, v < ps.max))
		b.WriteString("\n")
	}

	if len(le.themeKeys) > 0 {
		tkey := le.themeKeys[panels.ClampIndex(le.themeIdx, len(le.themeKeys))]
		b.WriteString(le.selectorRow(vk, "Theme", theme.DisplayName(tkey), le.cursor == le.themePos(), le.themeIdx > 0, le.themeIdx < len(le.themeKeys)-1))
		b.WriteString("\n")
	}
	if len(clockZoneChoices) > 0 {
		ci := panels.ClampIndex(le.clockIdx, len(clockZoneChoices))
		b.WriteString(le.selectorRow(vk, "World Clock", clockZoneChoices[ci].Label, le.cursor == le.clockPos(), ci > 0, ci < len(clockZoneChoices)-1))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	base := le.paneBase()
	for i, p := range es.panes {
		selected := le.cursor == base+i
		mark := theme.DimSty.Render("·")
		label := theme.DimSty.Render(p.title)
		if p.on {
			mark = theme.CanSty.Render("✓")
			label = theme.ValSty.Render(p.title)
			if selected {
				label = theme.AccentSty.Render(p.title)
			}
		} else if selected {
			label = theme.ValSty.Render(p.title)
		}
		var right string
		slimTag := "     "
		if p.slim {
			slimTag = theme.KeySty.Render("slim ")
		}
		if selected {
			up, down := " ", " "
			if i > 0 {
				up = theme.KeySty.Render("▴")
			}
			if i < len(es.panes)-1 {
				down = theme.KeySty.Render("▾")
			}
			right = up + down + "  " + slimTag + mark
		} else {
			right = slimTag + mark
		}
		b.WriteString(vk.Spread(vk.Selectable(label, selected), right))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.flash != "" {
		b.WriteString(panels.Flash(vk.Fit(m.flash)) + "\n\n")
	}

	km := layoutEditorKeymap()
	hints := [][2]string{km.Hint(keys.Up)}
	if le.cursor <= 1 {
		hints = append(hints, km.Hint(keys.Left))
	} else if le.cursor < base {
		hints = append(hints, km.HintLabeled(keys.Left, "adjust"))
	} else {
		hints = append(hints, km.Hint(keys.Confirm), km.Hint(keys.Inc), km.Hint(actLayoutSlim))
	}
	hints = append(hints, km.Hint(actLayoutSave), km.Hint(keys.Cancel))
	b.WriteString(vk.HintLine(hints...))
	return b.String()
}

func paramDisplay(ps paramSpec, v int) string {
	if ps.key == "rows" && v == 0 {
		return "auto"
	}
	return strconv.Itoa(v)
}

func (le *layoutEditorScreen) selectorRow(vk layout.Frame, label, value string, selected, canLeft, canRight bool) string {
	valSty := theme.ValSty
	if selected {
		valSty = theme.AccentSty
	}
	left, right := "  ", "  "
	if selected {
		if canLeft {
			left = theme.KeySty.Render("◂ ")
		}
		if canRight {
			right = theme.KeySty.Render(" ▸")
		}
	}
	return vk.Spread(vk.Selectable(label, selected), left+valSty.Render(value)+right)
}
