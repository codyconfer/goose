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
	return hint("tab/←/→/h/l", action)
}

func confirmHint(action string) [2]string {
	return hint("enter/space", action)
}
