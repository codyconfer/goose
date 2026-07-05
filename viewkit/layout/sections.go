package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

type FlexSections struct {
	MinWidth int
	MaxCols  int
}

func (g FlexSections) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	width := f.Width
	if width < 1 {
		width = theme.BodyWidth
	}

	visible := make([]Pane, 0, len(panes))
	for _, p := range panes {
		if tier >= p.MinTier {
			visible = append(visible, p)
		}
	}
	if len(visible) == 0 {
		return ""
	}

	order := make([]string, 0)
	groups := map[string][]Pane{}
	for _, p := range visible {
		if _, ok := groups[p.Group]; !ok {
			order = append(order, p.Group)
		}
		groups[p.Group] = append(groups[p.Group], p)
	}

	inner := FlexColumns{MinWidth: g.MinWidth, MaxCols: g.MaxCols}
	blocks := make([]string, 0, len(order))
	for _, name := range order {
		body := inner.Arrange(f, tier, groups[name], focusedName)
		if body == "" {
			continue
		}
		if name == "" {
			blocks = append(blocks, body)
			continue
		}
		blocks = append(blocks, StackTight(sectionHeader(width, name), body))
	}
	return Stack(blocks...)
}

func sectionHeader(width int, title string) string {
	label := theme.Cur().PanelTitle.Render(ansi.Truncate(title, width, "…"))
	rule := theme.Cur().Dim.Render(strings.Repeat("─", width))
	return label + "\n" + rule
}
