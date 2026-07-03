package keys

const (
	Up        Action = "nav.up"
	Down      Action = "nav.down"
	Left      Action = "nav.left"
	Right     Action = "nav.right"
	Confirm   Action = "confirm"
	Cancel    Action = "cancel"
	Quit      Action = "quit"
	FocusNext Action = "focus.next"
	FocusPrev Action = "focus.prev"
	Inc       Action = "inc"
	Dec       Action = "dec"
	Erase     Action = "erase"
	PageUp    Action = "page.up"
	PageDown  Action = "page.down"
)

type Scheme struct {
	bindings map[Action]Binding
}

func (s Scheme) Binding(a Action) Binding {
	return s.bindings[a]
}

func (s Scheme) With(overrides ...Binding) Scheme {
	next := Scheme{bindings: make(map[Action]Binding, len(s.bindings))}
	for k, v := range s.bindings {
		next.bindings[k] = v
	}
	for _, b := range overrides {
		next.bindings[b.Action] = b
	}
	return next
}

func Default() Scheme {
	return Scheme{bindings: map[Action]Binding{
		Up:        {Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k"},
		Down:      {Keys: []string{"down", "j"}, Action: Down},
		Left:      {Keys: []string{"left", "h"}, Action: Left, Glyph: "←/→/h/l"},
		Right:     {Keys: []string{"right", "l"}, Action: Right},
		Confirm:   {Keys: []string{"enter", " ", "spacebar"}, Action: Confirm, Glyph: "enter/space"},
		Cancel:    {Keys: []string{"esc"}, Action: Cancel},
		Quit:      {Keys: []string{"ctrl+c"}, Action: Quit},
		FocusNext: {Keys: []string{"tab"}, Action: FocusNext, Glyph: "tab/⇧tab", Label: "focus panel"},
		FocusPrev: {Keys: []string{"shift+tab"}, Action: FocusPrev},
		Inc:       {Keys: []string{"]", "+", "="}, Action: Inc, Glyph: "[ ]/-/+"},
		Dec:       {Keys: []string{"[", "-", "_"}, Action: Dec},
		Erase:     {Keys: []string{"backspace", "ctrl+h"}, Action: Erase, Glyph: "backspace"},
		PageUp:    {Keys: []string{"pgup"}, Action: PageUp},
		PageDown:  {Keys: []string{"pgdown"}, Action: PageDown},
	}}
}

var current = Default()

func Cur() Scheme { return current }

func Use(s Scheme) { current = s }
