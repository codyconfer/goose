package layout

import "github.com/codyconfer/viewkit/theme"

const (
	DefaultFlexMinWidth = 40
	DefaultFlexMaxCols  = 4
)

func FlexColumns(width, minWidth, maxCols int) int {
	if minWidth < 1 {
		minWidth = DefaultFlexMinWidth
	}
	if maxCols < 1 {
		maxCols = DefaultFlexMaxCols
	}
	cols := width / minWidth
	if cols < 1 {
		cols = 1
	}
	if cols > maxCols {
		cols = maxCols
	}
	return cols
}

type FlexGrid struct {
	MinWidth int
	MaxCols  int
}

func (g FlexGrid) Columns(width int) int {
	return FlexColumns(width, g.MinWidth, g.MaxCols)
}

func (g FlexGrid) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
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

	cols := g.Columns(width)
	if cols > len(visible) {
		cols = len(visible)
	}

	columns := make([][]Pane, cols)
	for i, p := range visible {
		columns[i%cols] = append(columns[i%cols], p)
	}

	colStr := make([]string, cols)
	maxH := 0
	for c := 0; c < cols; c++ {
		x := c * width / cols
		xEnd := (c + 1) * width / cols
		cw := xEnd - x
		sections := make([]Section, 0, len(columns[c]))
		for _, p := range columns[c] {
			pf := Frame{Width: cw}
			if p.Interactive && p.Name != "" && p.Name == focusedName {
				pf.Focused = true
			}
			sections = append(sections, Section{Content: p.Render(pf), MinTier: p.MinTier})
		}
		colStr[c] = StackFit(tier, sections...)
		if n := CountLines(colStr[c]); n > maxH {
			maxH = n
		}
	}
	if maxH < 1 {
		maxH = 1
	}

	rects := make([]rect, cols)
	for c := 0; c < cols; c++ {
		x := c * width / cols
		xEnd := (c + 1) * width / cols
		rects[c] = rect{x: x, y: 0, w: xEnd - x, h: maxH}
	}
	return composite(width, maxH, rects, colStr)
}
