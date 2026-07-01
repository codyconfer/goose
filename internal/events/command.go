package events

type Effect int

const (
	EffectFire Effect = iota
	EffectUnfire
)

type Command struct {
	Effect Effect `json:"effect"`
	Key    string `json:"key"`
}

func Fire(key string) Command { return Command{Effect: EffectFire, Key: key} }

func Unfire(key string) Command { return Command{Effect: EffectUnfire, Key: key} }

func (m *Machine) Apply(cmds []Command) {
	for _, c := range cmds {
		switch c.Effect {
		case EffectFire:
			m.markFired(c.Key)
		case EffectUnfire:
			m.clearFired(c.Key)
		}
	}
}

func (m *Machine) markFired(key string) {
	if m.s.Fired == nil {
		m.s.Fired = map[string]bool{}
	}
	m.s.Fired[key] = true
}

func (m *Machine) clearFired(key string) { delete(m.s.Fired, key) }
