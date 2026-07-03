package layout

type Focusable struct {
	Name        string
	Interactive bool
}

type Ring []string

func NewRing(all ...Focusable) Ring {
	out := make(Ring, 0, len(all))
	for _, f := range all {
		if f.Interactive {
			out = append(out, f.Name)
		}
	}
	return out
}

func (r Ring) At(idx int) string {
	if len(r) == 0 {
		return ""
	}
	return r[r.clamp(idx)]
}

func (r Ring) Step(idx, delta int) int {
	n := len(r)
	if n == 0 {
		return 0
	}
	return ((r.clamp(idx)+delta)%n + n) % n
}

func (r Ring) clamp(idx int) int {
	if idx < 0 {
		return 0
	}
	if idx >= len(r) {
		return len(r) - 1
	}
	return idx
}
