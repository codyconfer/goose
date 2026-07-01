package economy

func (s State) Frozen() bool { return s.FreezeSeconds > 0 }

func (m *Machine) freeze(seconds float64, reason string) {
	if seconds <= 0 {
		return
	}
	if seconds > m.s.FreezeSeconds {
		m.s.FreezeSeconds = seconds
		m.s.FreezeReason = reason
	}
}

func (m *Machine) TickFreeze(dt float64) bool {
	if m.s.FreezeSeconds <= 0 {
		return false
	}
	if dt > 0 {
		m.s.FreezeSeconds -= dt
		if m.s.FreezeSeconds <= 0 {
			m.s.FreezeSeconds = 0
			m.s.FreezeReason = ""
		}
	}
	return true
}
