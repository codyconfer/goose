package events

type State struct {
	Fired map[string]bool `json:"fired,omitempty"`
}

func NewState() State { return State{Fired: map[string]bool{}} }

func (s State) clone() State {
	out := State{Fired: make(map[string]bool, len(s.Fired))}
	for k, v := range s.Fired {
		out.Fired[k] = v
	}
	return out
}

func (s State) HasFired(key string) bool { return s.Fired[key] }

func (m *Machine) Get() State { return m.s.clone() }
