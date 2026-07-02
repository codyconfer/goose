package economy

type Machine struct {
	s State
}

const offlineStepSeconds = 1.0

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

	var earned float64
	remaining := float64(away)
	for remaining > 0 {
		dt := offlineStepSeconds
		if remaining < dt {
			dt = remaining
		}
		before := m.s.TotalEarned
		m.Produce(dt)
		earned += m.s.TotalEarned - before
		m.UpdateConsumers(dt)
		m.RunMarket(dt)
		m.ProcessTransactions(dt)
		m.UpdatePrice(dt, nil)
		m.TickPositions(dt)
		remaining -= dt
	}
	return earned
}
