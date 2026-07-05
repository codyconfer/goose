package layout

type Pane struct {
	Name        string
	Title       string
	Group       string
	Interactive bool
	MinTier     Tier
	Pos         *GridPos
	Slim        bool
	Render      func(Frame) string
}

func PaneRing(panes []Pane) Ring {
	fs := make([]Focusable, len(panes))
	for i, p := range panes {
		fs[i] = Focusable{Name: p.Name, Interactive: p.Interactive}
	}
	return NewRing(fs...)
}

type Layout interface {
	Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string
}

type Screen struct {
	Layout Layout
	Panes  []Pane
}

func (s Screen) Ring() Ring { return PaneRing(s.Panes) }

func (s Screen) Render(f Frame, tier Tier, focus int) string {
	l := s.Layout
	if l == nil {
		l = SingleColumn{}
	}
	return l.Arrange(f, tier, s.Panes, s.Ring().At(focus))
}

type SingleColumn struct{}

func (SingleColumn) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	sections := make([]Section, 0, len(panes))
	for _, p := range panes {
		pf := f
		if p.Interactive && p.Name != "" && p.Name == focusedName {
			pf = f.Focus()
		}
		sections = append(sections, Section{Content: p.Render(pf), MinTier: p.MinTier})
	}
	return StackFit(tier, sections...)
}
