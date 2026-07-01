package economy

type TxKind string

const (
	TxBuyEggs      TxKind = "buy-eggs"
	TxSellEggs     TxKind = "sell-eggs"
	TxBuyProducer  TxKind = "buy-producer"
	TxSellProducer TxKind = "sell-producer"
	TxUpgrade      TxKind = "upgrade"
	TxWindfall     TxKind = "windfall"
	TxOptionOpen   TxKind = "option-open"
	TxOptionSettle TxKind = "option-settle"
)

func (k TxKind) isEggTrade() bool { return k == TxBuyEggs || k == TxSellEggs }

type Transaction struct {
	Kind   TxKind  `json:"kind"`
	Label  string  `json:"label,omitempty"`
	Amount float64 `json:"amount"`
	Filled float64 `json:"filled,omitempty"`
	Tokens float64 `json:"tokens,omitempty"`
}

func (t Transaction) Remaining() float64 {
	if t.Filled >= t.Amount {
		return 0
	}
	return t.Amount - t.Filled
}

type TxReport struct {
	BoughtEggs   float64
	SoldEggs     float64
	TokensSpent  float64
	TokensEarned float64
	Completed    []Transaction
}

func (m *Machine) ScheduleTrade(kind TxKind, amount float64) bool {
	if amount <= 0 || !kind.isEggTrade() {
		return false
	}
	m.s.Transactions = append(m.s.Transactions, Transaction{Kind: kind, Amount: amount})
	return true
}

func (m *Machine) CancelTransaction(i int) bool {
	if i < 0 || i >= len(m.s.Transactions) {
		return false
	}
	m.s.Transactions = append(m.s.Transactions[:i], m.s.Transactions[i+1:]...)
	return true
}

func (m *Machine) ClearTransactions() { m.s.Transactions = nil }

func (m *Machine) fillTrade(kind TxKind, want, price float64) (eggs, tokens float64) {
	if want <= 0 || price <= 0 {
		return 0, 0
	}
	move := want
	switch kind {
	case TxBuyEggs:
		if aff := m.s.Tokens / price; move > aff {
			move = aff
		}
		if move <= 0 {
			return 0, 0
		}
		m.s.Tokens -= move * price
		m.gainEggs(move)
		m.s.EggsBought += move
	case TxSellEggs:
		if move > m.s.Eggs {
			move = m.s.Eggs
		}
		if move <= 0 {
			return 0, 0
		}
		m.s.Eggs -= move
		m.s.Tokens += move * price
	default:
		return 0, 0
	}
	return move, move * price
}

func (m *Machine) trade(kind TxKind, want, price float64) {
	m.fillTrade(kind, want, price)
}

func (s State) orderPrice(kind TxKind) float64 {
	if kind == TxSellEggs {
		return s.SellPrice()
	}
	return s.BuyPrice()
}

func (m *Machine) RunMarket(dt float64) (soldEggs, earned float64) {
	if dt <= 0 {
		return 0, 0
	}
	return m.fillTrade(TxSellEggs, m.s.Demand()*dt, m.s.SellPrice())
}

func (m *Machine) ProcessTransactions(dt float64) TxReport {
	var rep TxReport
	if len(m.s.Transactions) == 0 || dt <= 0 {
		return rep
	}
	lim := (m.s.Demand() + tradeFloorRate) * dt

	kept := m.s.Transactions[:0]
	for _, o := range m.s.Transactions {
		want := o.Remaining()
		if want > lim {
			want = lim
		}
		eggs, tokens := m.fillTrade(o.Kind, want, m.s.orderPrice(o.Kind))
		if eggs > 0 {
			o.Filled += eggs
			if o.Kind == TxBuyEggs {
				o.Tokens -= tokens
				rep.BoughtEggs += eggs
				rep.TokensSpent += tokens
			} else {
				o.Tokens += tokens
				rep.SoldEggs += eggs
				rep.TokensEarned += tokens
			}
		}
		if o.Remaining() <= tradeEpsilon {
			rep.Completed = append(rep.Completed, o)
			m.record(o)
			continue
		}
		kept = append(kept, o)
	}
	m.s.Transactions = kept
	return rep
}
