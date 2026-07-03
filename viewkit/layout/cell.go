package layout

func (f Frame) CellBox(title string, lines ...string) string {
	inner := NewFrame(f.Width - 4)
	if f.Focused {
		inner = inner.Focus()
	}
	box := inner.Panel(title, lines...)
	if f.Height > 0 {
		return FitBlock(box, f.Width, f.Height)
	}
	return box
}

func (f Frame) CellPanel(title string, lines []string, offset int) string {
	inner := NewFrame(f.Width - 4)
	if f.Focused {
		inner = inner.Focus()
	}
	rows := f.Height - 3
	if rows < 1 {
		rows = 1
	}
	box := inner.ScrollPanel(title, lines, rows, offset)
	if f.Height > 0 {
		return FitBlock(box, f.Width, f.Height)
	}
	return box
}
