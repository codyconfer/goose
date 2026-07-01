package economy

type Effect int

const (
	EffectEarn Effect = iota
	EffectSpend
	EffectGrantProducer
	EffectGrowCrowd
	EffectAddConsumers
	EffectTrade
	EffectSeizeBest
	EffectFreeze
	EffectMarginCall
	EffectShockPrice
)

type Command struct {
	Effect Effect
	Amount float64
	Factor float64
	Key    string
	Count  int
	Kind   TxKind
	Price  float64
}

func Earn(n float64) Command { return Command{Effect: EffectEarn, Amount: n} }

func Spend(n float64) Command { return Command{Effect: EffectSpend, Amount: n} }

func GrantProducer(key string, n int) Command {
	return Command{Effect: EffectGrantProducer, Key: key, Count: n}
}

func GrowCrowd(factor float64) Command { return Command{Effect: EffectGrowCrowd, Factor: factor} }

func AddConsumers(n float64) Command { return Command{Effect: EffectAddConsumers, Amount: n} }

func Trade(kind TxKind, eggs, price float64) Command {
	return Command{Effect: EffectTrade, Kind: kind, Amount: eggs, Price: price}
}

func SeizeBest() Command { return Command{Effect: EffectSeizeBest} }

func Freeze(seconds float64, reason string) Command {
	return Command{Effect: EffectFreeze, Amount: seconds, Key: reason}
}

func MarginCall(penalty float64) Command {
	return Command{Effect: EffectMarginCall, Factor: penalty}
}

func ShockPrice(factor float64) Command {
	return Command{Effect: EffectShockPrice, Factor: factor}
}

func (m *Machine) apply(cmds []Command) {
	for _, c := range cmds {
		switch c.Effect {
		case EffectEarn:
			m.earn(c.Amount)
		case EffectSpend:
			m.spend(c.Amount)
		case EffectGrantProducer:
			m.grantProducer(c.Key, c.Count)
		case EffectGrowCrowd:
			m.growCrowd(c.Factor)
		case EffectAddConsumers:
			m.addConsumers(c.Amount)
		case EffectTrade:
			m.trade(c.Kind, c.Amount, c.Price)
		case EffectSeizeBest:
			m.seizeBest()
		case EffectFreeze:
			m.freeze(c.Amount, c.Key)
		case EffectMarginCall:
			m.marginCallAll(c.Factor)
		case EffectShockPrice:
			m.shockPrice(c.Factor)
		}
	}
}

func (m *Machine) ApplyWindfall(label string, cmds []Command) {
	before := m.s.Tokens
	m.apply(cmds)
	m.recordWindfall(label, m.s.Tokens-before)
}
