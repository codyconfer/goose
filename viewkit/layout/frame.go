package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

type Frame struct {
	Width   int
	Height  int
	Focused bool
}

func NewFrame(width int) Frame {
	if width <= 0 {
		width = theme.BodyWidth
	}
	if width < theme.MinBodyWidth {
		width = theme.MinBodyWidth
	}
	return Frame{Width: width}
}

func (f Frame) Focus() Frame {
	f.Focused = true
	return f
}

func (f Frame) WithHeight(h int) Frame {
	f.Height = h
	return f
}

func DefaultFrame() Frame { return NewFrame(theme.BodyWidth) }

func (f Frame) BodyWidth() int {
	return NewFrame(f.Width).Width
}

func Spread(left, right string, width int) string {
	if width <= 0 {
		width = theme.BodyWidth
	}
	leftW, rightW := ansi.StringWidth(left), ansi.StringWidth(right)
	if leftW+rightW+1 > width {
		switch {
		case rightW >= width:
			return ansi.Truncate(right, width, "…")
		case width-rightW > 1:
			left = ansi.Truncate(left, width-rightW-1, "…")
		default:
			left = ""
		}
		leftW = ansi.StringWidth(left)
	}
	gap := max(width-leftW-rightW, 1)
	return left + strings.Repeat(" ", gap) + right
}

func (f Frame) Spread(left, right string) string {
	return Spread(left, right, f.BodyWidth())
}

func Fit(s string, width int) string {
	if width < 1 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

func (f Frame) Fit(s string) string {
	return Fit(s, f.BodyWidth())
}

func Rule() string {
	return DefaultFrame().Rule()
}

func (f Frame) Rule() string {
	return theme.Cur().Dim.Render(strings.Repeat("─", f.BodyWidth()+4))
}

func Header(title string, detail ...string) string {
	return DefaultFrame().Header(title, detail...)
}

func (f Frame) Header(title string, detail ...string) string {
	var head strings.Builder
	head.WriteString(theme.Cur().Title.Render(title))
	for _, part := range detail {
		if strings.TrimSpace(part) == "" {
			continue
		}
		head.WriteString(theme.Cur().Dim.Render("   ·   " + part))
	}
	return ansi.Truncate(head.String(), f.BodyWidth()+4, "…") + "\n" + f.Rule()
}

func Stack(sections ...string) string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		if section != "" {
			out = append(out, section)
		}
	}
	return strings.Join(out, "\n\n")
}

func StackTight(sections ...string) string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		if section != "" {
			out = append(out, section)
		}
	}
	return strings.Join(out, "\n")
}

func Box(lines ...string) string {
	return DefaultFrame().Box(lines...)
}

func (f Frame) Box(lines ...string) string {
	sty := theme.Cur().Panel
	if f.Focused {
		sty = theme.Cur().PanelFocus
	}
	return sty.Width(f.BodyWidth() + 2).Render(strings.Join(lines, "\n"))
}

func Panel(title string, lines ...string) string {
	return DefaultFrame().Panel(title, lines...)
}

func (f Frame) Panel(title string, lines ...string) string {
	return f.Box(append([]string{theme.Cur().PanelTitle.Render(ansi.Truncate(title, f.BodyWidth(), "…"))}, lines...)...)
}

func Row(label, value string) string {
	return DefaultFrame().Row(label, value)
}

func (f Frame) Row(label, value string) string {
	return f.Spread(theme.Cur().Dim.Render(label), value)
}

func HintLine(pairs ...[2]string) string {
	return DefaultFrame().HintLine(pairs...)
}

func (f Frame) HintLine(pairs ...[2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = theme.Cur().Key.Render(p[0]) + theme.Cur().Dim.Render(" "+p[1])
	}
	sep := theme.Cur().Dim.Render("   ·   ")
	var lines []string
	var line string
	for _, part := range parts {
		if line == "" {
			line = part
			continue
		}
		next := line + sep + part
		if ansi.StringWidth(next) <= f.BodyWidth() {
			line = next
			continue
		}
		lines = append(lines, line)
		line = part
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func Cursor(selected bool) string {
	if selected {
		return theme.Cur().Title.Render("▸ ")
	}
	return "  "
}

func Selectable(label string, selected bool) string {
	return DefaultFrame().Selectable(label, selected)
}

func (f Frame) Selectable(label string, selected bool) string {
	sty := theme.Cur().Val
	if selected {
		sty = theme.Cur().Accent
	}
	return Cursor(selected) + sty.Render(ansi.Truncate(label, f.BodyWidth()-2, "…"))
}
