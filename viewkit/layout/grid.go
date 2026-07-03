package layout

import (
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

type GridPos struct {
	Col     int
	Row     int
	ColSpan int
	RowSpan int
}

type Grid struct {
	Cols int
	Rows int
}

type gridCell struct {
	x, y, w, h int
}

type rect struct {
	x, y, w, h int
}

func (g Grid) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	height := f.Height
	if height < 1 {
		return SingleColumn{}.Arrange(f, tier, panes, focusedName)
	}
	visible := make([]Pane, 0, len(panes))
	for _, p := range panes {
		if tier >= p.MinTier {
			visible = append(visible, p)
		}
	}
	if len(visible) == 0 {
		return strings.Repeat("\n", height-1)
	}

	width := f.Width
	if width < 1 {
		width = theme.BodyWidth
	}
	cells, cols, rows := g.place(visible)
	rects := make([]rect, len(visible))
	blocks := make([]string, len(visible))
	for i, p := range visible {
		r := pixelRect(cells[i], cols, rows, width, height)
		rects[i] = r
		pf := Frame{Width: r.w, Height: r.h}
		if p.Interactive && p.Name != "" && p.Name == focusedName {
			pf.Focused = true
		}
		blocks[i] = p.Render(pf)
	}
	return composite(width, height, rects, blocks)
}

func (g Grid) place(panes []Pane) (cells []gridCell, cols, rows int) {
	cols = g.Cols
	if cols < 1 {
		cols = 1
	}
	cells = make([]gridCell, len(panes))
	occupied := map[[2]int]bool{}
	for i := range panes {
		if panes[i].Pos == nil {
			continue
		}
		c := cellFor(*panes[i].Pos, cols)
		cells[i] = c
		occupy(occupied, c)
	}
	for i := range panes {
		if panes[i].Pos != nil {
			continue
		}
		x, y := nextFree(occupied, cols)
		c := gridCell{x: x, y: y, w: 1, h: 1}
		cells[i] = c
		occupy(occupied, c)
	}
	rows = g.Rows
	for _, c := range cells {
		if c.y+c.h > rows {
			rows = c.y + c.h
		}
	}
	if rows < 1 {
		rows = 1
	}
	return cells, cols, rows
}

func cellFor(p GridPos, cols int) gridCell {
	w := p.ColSpan
	if w < 1 {
		w = 1
	}
	h := p.RowSpan
	if h < 1 {
		h = 1
	}
	x := p.Col
	if x < 0 {
		x = 0
	}
	if x > cols-1 {
		x = cols - 1
	}
	if x+w > cols {
		w = cols - x
	}
	if w < 1 {
		w = 1
	}
	y := p.Row
	if y < 0 {
		y = 0
	}
	return gridCell{x: x, y: y, w: w, h: h}
}

func occupy(occ map[[2]int]bool, c gridCell) {
	for dy := 0; dy < c.h; dy++ {
		for dx := 0; dx < c.w; dx++ {
			occ[[2]int{c.x + dx, c.y + dy}] = true
		}
	}
}

func nextFree(occ map[[2]int]bool, cols int) (int, int) {
	for y := 0; ; y++ {
		for x := 0; x < cols; x++ {
			if !occ[[2]int{x, y}] {
				return x, y
			}
		}
	}
}

func pixelRect(c gridCell, cols, rows, width, height int) rect {
	x := c.x * width / cols
	xEnd := (c.x + c.w) * width / cols
	y := c.y * height / rows
	yEnd := (c.y + c.h) * height / rows
	return rect{x: x, y: y, w: xEnd - x, h: yEnd - y}
}

type segment struct {
	x    int
	text string
}

func composite(width, height int, rects []rect, blocks []string) string {
	rowSegs := make([][]segment, height)
	for i, r := range rects {
		lines := fitLines(blocks[i], r.w, r.h)
		for dy := 0; dy < r.h; dy++ {
			ry := r.y + dy
			if ry < 0 || ry >= height {
				continue
			}
			rowSegs[ry] = append(rowSegs[ry], segment{x: r.x, text: lines[dy]})
		}
	}
	out := make([]string, height)
	for y := 0; y < height; y++ {
		segs := rowSegs[y]
		sort.SliceStable(segs, func(a, b int) bool { return segs[a].x < segs[b].x })
		var b strings.Builder
		cursor := 0
		for _, s := range segs {
			if s.x > cursor {
				b.WriteString(strings.Repeat(" ", s.x-cursor))
				cursor = s.x
			}
			b.WriteString(s.text)
			cursor += ansi.StringWidth(s.text)
		}
		out[y] = b.String()
	}
	return strings.Join(out, "\n")
}

func FitBlock(block string, w, h int) string {
	return strings.Join(fitLines(block, w, h), "\n")
}

func fitLines(block string, w, h int) []string {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	raw := strings.Split(block, "\n")
	out := make([]string, h)
	blank := strings.Repeat(" ", w)
	for i := 0; i < h; i++ {
		if i < len(raw) {
			out[i] = padTo(raw[i], w)
		} else {
			out[i] = blank
		}
	}
	return out
}

func padTo(line string, w int) string {
	lw := ansi.StringWidth(line)
	switch {
	case lw > w:
		return ansi.Truncate(line, w, "")
	case lw < w:
		return line + strings.Repeat(" ", w-lw)
	default:
		return line
	}
}
