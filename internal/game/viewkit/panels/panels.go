package panels

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type Frame struct {
	Width   int
	Focused bool // draw the panel border highlighted to mark keyboard focus
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

// Focus returns a copy of the frame that renders panels with the focused border.
func (f Frame) Focus() Frame {
	f.Focused = true
	return f
}

func DefaultFrame() Frame { return NewFrame(theme.BodyWidth) }

func (f Frame) bodyWidth() int {
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
	return Spread(left, right, f.bodyWidth())
}

func Fit(s string, width int) string {
	if width < 1 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

func (f Frame) Fit(s string) string {
	return Fit(s, f.bodyWidth())
}

func Rule() string {
	return DefaultFrame().Rule()
}

func (f Frame) Rule() string {
	return theme.DimSty.Render(strings.Repeat("─", f.bodyWidth()+4))
}

func Header(title string, detail ...string) string {
	return DefaultFrame().Header(title, detail...)
}

func (f Frame) Header(title string, detail ...string) string {
	var head strings.Builder
	head.WriteString(theme.TitleSty.Render(title))
	for _, part := range detail {
		if strings.TrimSpace(part) == "" {
			continue
		}
		head.WriteString(theme.DimSty.Render("   ·   " + part))
	}
	return ansi.Truncate(head.String(), f.bodyWidth()+4, "…") + "\n" + f.Rule()
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

func Box(lines ...string) string {
	return DefaultFrame().Box(lines...)
}

func (f Frame) Box(lines ...string) string {
	sty := theme.PanelSty
	if f.Focused {
		sty = theme.PanelFocusSty
	}
	return sty.Width(f.bodyWidth() + 2).Render(strings.Join(lines, "\n"))
}

func Panel(title string, lines ...string) string {
	return DefaultFrame().Panel(title, lines...)
}

func (f Frame) Panel(title string, lines ...string) string {
	return f.Box(append([]string{theme.PanelTitleSty.Render(ansi.Truncate(title, f.bodyWidth(), "…"))}, lines...)...)
}

func Row(label, value string) string {
	return DefaultFrame().Row(label, value)
}

func (f Frame) Row(label, value string) string {
	return f.Spread(theme.DimSty.Render(label), value)
}

func ProgressBar(frac float64, width int) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac * float64(width))
	return theme.EggSty.Render(strings.Repeat("█", filled)) + theme.DimSty.Render(strings.Repeat("░", width-filled))
}

func Meter(frac float64, width int) string {
	return "[" + ProgressBar(frac, width) + "]"
}

func HintLine(pairs ...[2]string) string {
	return DefaultFrame().HintLine(pairs...)
}

func (f Frame) HintLine(pairs ...[2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = theme.KeySty.Render(p[0]) + theme.DimSty.Render(" "+p[1])
	}
	sep := theme.DimSty.Render("   ·   ")
	var lines []string
	var line string
	for _, part := range parts {
		if line == "" {
			line = part
			continue
		}
		next := line + sep + part
		if ansi.StringWidth(next) <= f.bodyWidth() {
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

func Flash(message string) string {
	if message == "" {
		return ""
	}
	return theme.DimSty.Italic(true).Render(message)
}

func Cursor(selected bool) string {
	if selected {
		return theme.TitleSty.Render("▸ ")
	}
	return "  "
}

func Selectable(label string, selected bool) string {
	return DefaultFrame().Selectable(label, selected)
}

func (f Frame) Selectable(label string, selected bool) string {
	sty := theme.ValSty
	if selected {
		sty = theme.EggSty
	}
	return Cursor(selected) + sty.Render(ansi.Truncate(label, f.bodyWidth()-2, "…"))
}

func Toggle(left, right string, leftActive bool) string {
	leftSty, rightSty := theme.ValSty, theme.ValSty
	if leftActive {
		leftSty = theme.EggSty
	} else {
		rightSty = theme.EggSty
	}
	return leftSty.Render(left) + theme.DimSty.Render("  /  ") + rightSty.Render(right)
}

func ClampIndex(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= total {
		return total - 1
	}
	return index
}

func MoveIndex(index, delta, total int) int {
	return ClampIndex(index+delta, total)
}

func StepIndex(index, delta, total int) int {
	return MoveIndex(index, delta, total)
}
