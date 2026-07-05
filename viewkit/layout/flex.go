package layout

import "github.com/codyconfer/viewkit/theme"

const (
	DefaultFlexMinWidth = 40
	DefaultFlexMaxCols  = 3
)

// FlexColCount returns how many equal-width tracks fit in width, given a
// minimum track width and a hard cap, clamped to [1, maxCols].
func FlexColCount(width, minWidth, maxCols int) int {
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

// FlexColumns is a responsive column-masonry layout: panes are distributed
// round-robin into N equal-width vertical stacks, where N grows with width.
// Columns keep a fixed width, so a pane never expands into an empty neighbour.
type FlexColumns struct {
	MinWidth int
	MaxCols  int
}

func (g FlexColumns) Columns(width int) int {
	return FlexColCount(width, g.MinWidth, g.MaxCols)
}

func (g FlexColumns) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
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

// FlexRows is a responsive row-major layout: panes fill left-to-right into
// rows of up to N tracks (N grows with width). A row with fewer panes than N
// — most often the ragged final row — divides the full width among only the
// panes present, so a pane with no neighbour expands to fill the row. Each row
// is as tall as its tallest pane; shorter panes are padded to match.
type FlexRows struct {
	MinWidth int
	MaxCols  int
}

func (g FlexRows) Columns(width int) int {
	return FlexColCount(width, g.MinWidth, g.MaxCols)
}

func (g FlexRows) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
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

	rows := make([]string, 0, (len(visible)+cols-1)/cols)
	for start := 0; start < len(visible); start += cols {
		end := start + cols
		if end > len(visible) {
			end = len(visible)
		}
		row := visible[start:end]
		k := len(row)

		rects := make([]rect, k)
		blocks := make([]string, k)
		rowH := 1
		for j, p := range row {
			x := j * width / k
			xEnd := (j + 1) * width / k
			pf := Frame{Width: xEnd - x}
			if p.Interactive && p.Name != "" && p.Name == focusedName {
				pf.Focused = true
			}
			blocks[j] = p.Render(pf)
			rects[j] = rect{x: x, y: 0, w: xEnd - x}
			if n := CountLines(blocks[j]); n > rowH {
				rowH = n
			}
		}
		for j := range rects {
			rects[j].h = rowH
		}
		rows = append(rows, composite(width, rowH, rects, blocks))
	}
	return Stack(rows...)
}
