package economy

type Machine struct {
	s State
}

func NewMachine() *Machine { return &Machine{s: NewState()} }

func FromState(s State) *Machine {
	Normalize(&s)
	return &Machine{s: s}
}

func (m *Machine) SetLastSeen(unix int64) { m.s.LastSeen = unix }

func (m *Machine) LastSeen() int64 { return m.s.LastSeen }

func (m *Machine) Tap() {
	if m.s.Frozen() {
		return
	}
	m.earn(m.s.PerClick)
}

func (m *Machine) earn(n float64) {
	m.s.Tokens += n
	m.s.TotalEarned += n
}

func (m *Machine) spend(n float64) {
	if m.s.Tokens -= n; m.s.Tokens < 0 {
		m.s.Tokens = 0
	}
}

func (m *Machine) gainEggs(n float64) {
	m.s.Eggs += n
	if m.s.Eggs > m.s.PeakEggs {
		m.s.PeakEggs = m.s.Eggs
	}
}

func (m *Machine) layEggs(n float64) {
	if n <= 0 {
		return
	}
	m.s.EggsLaid += n
	m.gainEggs(n)
}

func (m *Machine) Produce(dt float64) {
	if m.s.Frozen() {
		return
	}
	if gain := m.s.TokensPerSecond() * dt; gain > 0 {
		m.earn(gain)
	}
	if eggs := m.s.EggsPerSecond() * dt; eggs > 0 {
		m.layEggs(eggs)
	}
}

func (m *Machine) ApplyOffline(away int64) float64 {
	if away <= 0 {
		return 0
	}
	if away > maxOfflineSeconds {
		away = maxOfflineSeconds
	}

	if m.s.FreezeSeconds > 0 {
		frozen := m.s.FreezeSeconds
		if frozen > float64(away) {
			frozen = float64(away)
		}
		m.s.FreezeSeconds -= frozen
		if m.s.FreezeSeconds <= 0 {
			m.s.FreezeSeconds = 0
			m.s.FreezeReason = ""
		}
		if away -= int64(frozen); away <= 0 {
			return 0
		}
	}

	earned := m.s.TokensPerSecond() * float64(away)
	m.earn(earned)

	m.layEggs(m.s.EggsPerSecond() * float64(away))

	m.RunMarket(float64(away))
	m.TickPositions(float64(away))
	return earned
}
