package game

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
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
	"sections":     "Sections",
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
	case "flex-columns", "flex-rows", "sections":
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
	list      list.Model
	ready     bool
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

func (le *layoutEditorScreen) ensure() {
	if le.ready {
		return
	}
	le.list = list.New()
	le.list.SetFocused(true)
	le.sync(layout.DefaultFrame(), "")
	le.ready = true
}

func (le *layoutEditorScreen) sync(vk layout.Frame, sel string) {
	rows := le.rows(vk, sel)
	items := make([]list.Item, len(rows))
	for i, r := range rows {
		items[i] = list.Item{Block: r.block, Key: r.key, Selectable: true}
	}
	le.list.SetItems(items)
	if sel != "" {
		selectListKey(&le.list, sel)
	}
}

func (le *layoutEditorScreen) selKey() string {
	if it, ok := le.list.Selected(); ok {
		return it.Key
	}
	return ""
}

func (le *layoutEditorScreen) selectedPaneKey() (string, bool) {
	if key := le.selKey(); strings.HasPrefix(key, "pane:") {
		return strings.TrimPrefix(key, "pane:"), true
	}
	return "", false
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

func (le *layoutEditorScreen) togglePane(key string, fn func(*editPane)) {
	es := le.current()
	for i := range es.panes {
		if es.panes[i].key == key {
			fn(&es.panes[i])
			return
		}
	}
}

func (le *layoutEditorScreen) stepParam(pk string, delta int) {
	for _, ps := range le.paramSpecs() {
		if ps.key != pk {
			continue
		}
		v := le.paramValue(ps) + delta
		if v < ps.min {
			v = ps.min
		}
		if v > ps.max {
			v = ps.max
		}
		le.current().params[ps.key] = v
		return
	}
}

func (le *layoutEditorScreen) paramValue(ps paramSpec) int {
	if v, ok := le.current().params[ps.key]; ok {
		return v
	}
	return ps.def
}

func (le *layoutEditorScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	le.ensure()
	action, ok := layoutEditorKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		return tea.Quit
	case keys.Cancel:
		m.screen = le.prev
	case keys.Up:
		le.list.Move(-1)
	case keys.Down:
		le.list.Move(1)
	case keys.Left:
		le.changeRow(-1)
	case keys.Right:
		le.changeRow(1)
	case keys.Confirm:
		if key, ok := le.selectedPaneKey(); ok {
			le.togglePane(key, func(p *editPane) { p.on = !p.on })
		}
	case actLayoutSlim:
		if key, ok := le.selectedPaneKey(); ok {
			le.togglePane(key, func(p *editPane) { p.slim = !p.slim })
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
	switch key := le.selKey(); {
	case key == "screen":
		le.screenIdx = panels.StepIndex(le.screenIdx, delta, len(configurableScreens))
	case key == "layout":
		es := le.current()
		if len(es.layouts) > 0 {
			es.layoutIdx = panels.StepIndex(es.layoutIdx, delta, len(es.layouts))
		}
	case key == "theme":
		le.stepTheme(delta)
	case key == "clock":
		le.stepClock(delta)
	case strings.HasPrefix(key, "param:"):
		le.stepParam(strings.TrimPrefix(key, "param:"), delta)
	}
}

func (le *layoutEditorScreen) reorder(delta int) {
	key, ok := le.selectedPaneKey()
	if !ok {
		return
	}
	es := le.current()
	p := -1
	for i := range es.panes {
		if es.panes[i].key == key {
			p = i
			break
		}
	}
	q := p + delta
	if p < 0 || q < 0 || q >= len(es.panes) {
		return
	}
	es.panes[p], es.panes[q] = es.panes[q], es.panes[p]
	le.sync(layout.DefaultFrame(), "pane:"+key)
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
	le.ensure()
	vk := m.frame()
	le.sync(vk, le.selKey())
	le.list.SetSize(vk.Width, 0)

	var b strings.Builder
	b.WriteString(vk.Header("SCREEN LAYOUT", "pick a layout, then toggle, reorder, and slim panels"))
	b.WriteString("\n\n")
	b.WriteString(le.list.View())
	b.WriteString("\n\n")

	if m.flash != "" {
		b.WriteString(panels.Flash(vk.Fit(m.flash)) + "\n\n")
	}

	km := layoutEditorKeymap()
	hints := [][2]string{km.Hint(keys.Up)}
	switch key := le.selKey(); {
	case strings.HasPrefix(key, "pane:"):
		hints = append(hints, km.Hint(keys.Confirm), km.Hint(keys.Inc), km.Hint(actLayoutSlim))
	case key == "screen" || key == "layout":
		hints = append(hints, km.Hint(keys.Left))
	default:
		hints = append(hints, km.HintLabeled(keys.Left, "adjust"))
	}
	hints = append(hints, km.Hint(actLayoutSave), km.Hint(keys.Cancel))
	b.WriteString(vk.HintLine(hints...))
	return b.String()
}

type editRow struct {
	key   string
	block string
}

func (le *layoutEditorScreen) rows(vk layout.Frame, sel string) []editRow {
	id := configurableScreens[le.screenIdx]
	es := le.current()
	width := vk.Width - 2

	rows := []editRow{
		{"screen", rowLine(width, "Screen", screenTitles[id], sel == "screen")},
		{"layout", rowLine(width, "Layout", layoutDisplayName(le.currentLayoutKey()), sel == "layout")},
	}
	for _, ps := range le.paramSpecs() {
		key := "param:" + ps.key
		rows = append(rows, editRow{key, rowLine(width, ps.label, paramDisplay(ps, le.paramValue(ps)), sel == key)})
	}
	if len(le.themeKeys) > 0 {
		tkey := le.themeKeys[panels.ClampIndex(le.themeIdx, len(le.themeKeys))]
		rows = append(rows, editRow{"theme", rowLine(width, "Theme", theme.DisplayName(tkey), sel == "theme")})
	}
	if len(clockZoneChoices) > 0 {
		ci := panels.ClampIndex(le.clockIdx, len(clockZoneChoices))
		rows = append(rows, editRow{"clock", rowLine(width, "World Clock", clockZoneChoices[ci].Label, sel == "clock")})
	}
	for _, p := range es.panes {
		key := "pane:" + p.key
		rows = append(rows, editRow{key, paneLine(width, p, sel == key)})
	}
	return rows
}

func paramDisplay(ps paramSpec, v int) string {
	if ps.key == "rows" && v == 0 {
		return "auto"
	}
	return strconv.Itoa(v)
}

func rowLine(width int, label, value string, selected bool) string {
	sty := theme.ValSty
	if selected {
		sty = theme.AccentSty
	}
	return layout.Spread(sty.Render(label), sty.Render(value), width)
}

func paneLine(width int, p editPane, selected bool) string {
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
	slimTag := "     "
	if p.slim {
		slimTag = theme.KeySty.Render("slim ")
	}
	return layout.Spread(label, slimTag+mark, width)
}
