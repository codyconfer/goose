package game

import (
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
}

type editSpec struct {
	layoutIdx int
	layouts   []string
	panes     []editPane
}

type layoutEditorScreen struct {
	prev      screen
	screenIdx int
	cursor    int
	specs     map[string]*editSpec
}

func newLayoutEditor(prev screen) *layoutEditorScreen {
	specs := make(map[string]*editSpec, len(configurableScreens))
	for _, id := range configurableScreens {
		specs[id] = newEditSpec(id)
	}
	return &layoutEditorScreen{prev: prev, specs: specs}
}

func newEditSpec(id string) *editSpec {
	spec := layoutSpec(id)
	infos := paneRegistryKeys(id)
	title := make(map[string]string, len(infos))
	for _, in := range infos {
		title[in.Key] = in.Title
	}

	seen := map[string]bool{}
	var panes []editPane
	for _, ref := range spec.Panes {
		t, ok := title[ref.Key]
		if !ok {
			continue
		}
		panes = append(panes, editPane{key: ref.Key, title: t, on: true})
		seen[ref.Key] = true
	}
	for _, in := range infos {
		if seen[in.Key] {
			continue
		}
		panes = append(panes, editPane{key: in.Key, title: in.Title, on: false})
	}

	layouts := layoutKeys(id)
	return &editSpec{layoutIdx: layoutIndex(layouts, spec.Layout), layouts: layouts, panes: panes}
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

func (le *layoutEditorScreen) rowCount() int {
	return 2 + len(le.current().panes)
}

func (le *layoutEditorScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := layoutEditorKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	es := le.current()
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
		if p := le.cursor - 2; p >= 0 && p < len(es.panes) {
			es.panes[p].on = !es.panes[p].on
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
	switch le.cursor {
	case 0:
		le.screenIdx = panels.StepIndex(le.screenIdx, delta, len(configurableScreens))
		le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	case 1:
		es := le.current()
		if len(es.layouts) > 0 {
			es.layoutIdx = panels.StepIndex(es.layoutIdx, delta, len(es.layouts))
		}
	}
}

func (le *layoutEditorScreen) reorder(delta int) {
	es := le.current()
	p := le.cursor - 2
	q := p + delta
	if p < 0 || p >= len(es.panes) || q < 0 || q >= len(es.panes) {
		return
	}
	es.panes[p], es.panes[q] = es.panes[q], es.panes[p]
	le.cursor = q + 2
}

func (le *layoutEditorScreen) apply() error {
	for id, es := range le.specs {
		var refs []layout.PaneRef
		for _, p := range es.panes {
			if p.on {
				refs = append(refs, layout.PaneRef{Key: p.key})
			}
		}
		lay := "single"
		if es.layoutIdx >= 0 && es.layoutIdx < len(es.layouts) {
			lay = es.layouts[es.layoutIdx]
		}
		setLayoutSpec(id, layout.ScreenSpec{Layout: lay, Panes: refs})
	}
	return saveLayoutConfig()
}

func (le *layoutEditorScreen) view(m *Model) string {
	vk := m.frame()
	le.cursor = panels.ClampIndex(le.cursor, le.rowCount())
	id := configurableScreens[le.screenIdx]
	es := le.current()

	var b strings.Builder
	b.WriteString(vk.Header("SCREEN LAYOUT", "pick a layout, then toggle and reorder panels"))
	b.WriteString("\n\n")

	b.WriteString(le.selectorRow(vk, "Screen", screenTitles[id], le.cursor == 0, le.screenIdx > 0, le.screenIdx < len(configurableScreens)-1))
	b.WriteString("\n")
	layName := "single"
	if es.layoutIdx >= 0 && es.layoutIdx < len(es.layouts) {
		layName = es.layouts[es.layoutIdx]
	}
	b.WriteString(le.selectorRow(vk, "Layout", layName, le.cursor == 1, es.layoutIdx > 0, es.layoutIdx < len(es.layouts)-1))
	b.WriteString("\n\n")

	for i, p := range es.panes {
		selected := le.cursor == 2+i
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
		b.WriteString(vk.Spread(vk.Selectable(label, selected), mark))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.flash != "" {
		b.WriteString(panels.Flash(vk.Fit(m.flash)) + "\n\n")
	}

	km := layoutEditorKeymap()
	hints := [][2]string{km.Hint(keys.Up)}
	switch le.cursor {
	case 0, 1:
		hints = append(hints, km.Hint(keys.Left))
	default:
		hints = append(hints, km.Hint(keys.Confirm), km.Hint(keys.Inc))
	}
	hints = append(hints, km.Hint(actLayoutSave), km.Hint(keys.Cancel))
	b.WriteString(vk.HintLine(hints...))
	return b.String()
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
