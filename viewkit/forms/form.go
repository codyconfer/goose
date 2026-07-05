package forms

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
)

type Form struct {
	Fields []Field
	cursor int
}

func NewForm(fields ...Field) *Form {
	return &Form{Fields: fields}
}

func (fm *Form) Focused() *Field {
	if len(fm.Fields) == 0 {
		return nil
	}
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	return &fm.Fields[fm.cursor]
}

func (fm *Form) Handle(a keys.Action) bool {
	if len(fm.Fields) == 0 {
		return false
	}
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	fd := &fm.Fields[fm.cursor]

	switch a {
	case keys.Up, keys.FocusPrev:
		fm.cursor = panels.MoveIndex(fm.cursor, -1, len(fm.Fields))
	case keys.Down, keys.FocusNext:
		fm.cursor = panels.MoveIndex(fm.cursor, +1, len(fm.Fields))
	case keys.Left, keys.Dec:
		fd.left()
	case keys.Right, keys.Inc:
		fd.right()
	case keys.Erase:
		fd.backspace()
	case keys.Confirm:
		return fd.activate()
	default:
		return false
	}
	return true
}

func (fm *Form) Insert(s string) {
	if fd := fm.Focused(); fd != nil {
		fd.insert(s)
	}
}

func (fm *Form) Values() map[string]any {
	out := make(map[string]any, len(fm.Fields))
	for i := range fm.Fields {
		out[fm.Fields[i].Key] = fm.Fields[i].Value()
	}
	return out
}

func (fm *Form) Render(f layout.Frame, title string) string {
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	var lines []string
	for i := range fm.Fields {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, fm.Fields[i].render(f, i == fm.cursor)...)
	}
	return f.Panel(title, lines...)
}

func (fm *Form) Overlay(bg string, f layout.Frame, title string, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, fm.Render(f, title), pos...)
}
