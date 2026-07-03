package panels

import (
	"fmt"
	"strings"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type ScrollState struct {
	Offset int
}

func (s *ScrollState) Scroll(delta, total, rows int) {
	s.Offset += delta
	s.clamp(total, rows)
}

func (s *ScrollState) Reveal(index, total, rows int) {
	if rows < 1 {
		rows = 1
	}
	if index < s.Offset {
		s.Offset = index
	} else if index >= s.Offset+rows {
		s.Offset = index - rows + 1
	}
	s.clamp(total, rows)
}

func (s *ScrollState) clamp(total, rows int) {
	max := total - rows
	if max < 0 {
		max = 0
	}
	if s.Offset > max {
		s.Offset = max
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
}

func scrollWindow(lines []string, rows, offset int) (window []string, footer string, ok bool) {
	total := len(lines)
	if rows < 1 {
		rows = 1
	}
	if total <= rows {
		return lines, "", false
	}
	max := total - rows
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + rows
	return lines[offset:end], fmt.Sprintf("↕ %d–%d of %d", offset+1, end, total), true
}

func ScrollPanel(title string, lines []string, rows, offset int) string {
	return DefaultFrame().ScrollPanel(title, lines, rows, offset)
}

func (f Frame) ScrollPanel(title string, lines []string, rows, offset int) string {
	return f.ScrollPanelWithPrefix(title, nil, lines, rows, offset)
}

func (f Frame) ScrollPanelWithPrefix(title string, prefix, lines []string, rows, offset int) string {
	if len(lines) == 0 {
		return f.Panel(title, prefix...)
	}
	window, footer, ok := scrollWindow(lines, rows, offset)
	out := make([]string, 0, len(prefix)+len(window)+1)
	out = append(out, prefix...)
	out = append(out, window...)
	if ok {
		out = append(out, theme.DimSty.Render(footer))
	}
	return f.Panel(title, out...)
}

func Viewport(body string, rows, offset int) string {
	lines := strings.Split(body, "\n")
	if rows < 1 {
		return ""
	}
	total := len(lines)
	if total <= rows {
		return body
	}
	if rows == 1 {
		off := clampOffset(offset, total, 1)
		return viewportHint(off, off+1, total)
	}

	windowRows := rows - 1
	off := clampOffset(offset, total, windowRows)
	end := off + windowRows
	out := make([]string, 0, windowRows+1)
	out = append(out, lines[off:end]...)
	out = append(out, viewportHint(off, end, total))
	return strings.Join(out, "\n")
}

func clampOffset(offset, total, rows int) int {
	max := total - rows
	if max < 0 {
		max = 0
	}
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}

// viewportHint renders the outer viewport's scroll indicator: directional
// arrows show whether content is hidden above/below, alongside the paging key
// and a position count.
func viewportHint(offset, end, total int) string {
	up, down := "  ", "  "
	if offset > 0 {
		up = "▲ "
	}
	if end < total {
		down = "▼ "
	}
	return theme.DimSty.Render(fmt.Sprintf("%s%s pgup/pgdn  ·  %d–%d of %d", up, down, offset+1, end, total))
}
