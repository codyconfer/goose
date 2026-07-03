package keys

import "strings"

type Action string

type Binding struct {
	Keys   []string
	Action Action
	Glyph  string
	Label  string
}

func (b Binding) DisplayGlyph() string {
	if b.Glyph != "" {
		return b.Glyph
	}
	return strings.Join(b.Keys, "/")
}

func (b Binding) WithGlyph(glyph string) Binding {
	b.Glyph = glyph
	return b
}

func (b Binding) WithLabel(label string) Binding {
	b.Label = label
	return b
}

type Map struct {
	byKey map[string]Action
	byAct map[Action]Binding
}

func NewMap(bindings ...Binding) *Map {
	m := &Map{
		byKey: make(map[string]Action, len(bindings)*2),
		byAct: make(map[Action]Binding, len(bindings)),
	}
	for _, b := range bindings {
		m.byAct[b.Action] = b
		for _, k := range b.Keys {
			m.byKey[k] = b.Action
		}
	}
	return m
}

func (m *Map) Action(input string) (a Action, ok bool) {
	a, ok = m.byKey[input]
	return a, ok
}

func (m *Map) Has(a Action) bool {
	_, ok := m.byAct[a]
	return ok
}

func (m *Map) Hint(a Action) [2]string {
	b := m.byAct[a]
	return [2]string{b.DisplayGlyph(), b.Label}
}

func (m *Map) HintLabeled(a Action, label string) [2]string {
	return [2]string{m.byAct[a].DisplayGlyph(), label}
}

func (m *Map) Hints(actions ...Action) [][2]string {
	out := make([][2]string, 0, len(actions))
	for _, a := range actions {
		if b, ok := m.byAct[a]; ok && b.Glyph != "" {
			out = append(out, [2]string{b.Glyph, b.Label})
		}
	}
	return out
}
