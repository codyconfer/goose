package game

type focusable struct {
	name        string
	interactive bool
}

func focusNames(all ...focusable) []string {
	out := make([]string, 0, len(all))
	for _, f := range all {
		if f.interactive {
			out = append(out, f.name)
		}
	}
	return out
}

func focusResolve(names []string, idx int) string {
	if len(names) == 0 {
		return ""
	}
	return names[clampFocus(names, idx)]
}

func focusStep(names []string, idx, delta int) int {
	n := len(names)
	if n == 0 {
		return 0
	}
	return ((clampFocus(names, idx)+delta)%n + n) % n
}

func clampFocus(names []string, idx int) int {
	if idx < 0 {
		return 0
	}
	if idx >= len(names) {
		return len(names) - 1
	}
	return idx
}
