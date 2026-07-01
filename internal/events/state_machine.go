package events

type Machine struct {
	s State
}

func NewMachine() *Machine { return &Machine{s: NewState()} }

func FromState(s State) *Machine {
	if s.Fired == nil {
		s.Fired = map[string]bool{}
	}
	return &Machine{s: s}
}
