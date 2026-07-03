package game

func hint(keys, action string) [2]string {
	return [2]string{keys, action}
}

func verticalHint(action string) [2]string {
	return hint("↑/↓/j/k", action)
}

func horizontalHint(action string) [2]string {
	return hint("←/→/h/l", action)
}

func toggleHint(action string) [2]string {
	return hint("←/→/h/l", action)
}

func confirmHint(action string) [2]string {
	return hint("enter/space", action)
}

func focusHints(verb string, ringSize int) [][2]string {
	hints := [][2]string{verticalHint(verb)}
	if ringSize > 1 {
		hints = append(hints, hint("tab/⇧tab", "focus panel"))
	}
	return hints
}
