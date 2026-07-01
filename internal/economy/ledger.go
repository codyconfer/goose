package economy

func (m *Machine) record(tx Transaction) {
	m.s.Ledger = append(m.s.Ledger, tx)
	if len(m.s.Ledger) > ledgerMax {
		m.s.Ledger = m.s.Ledger[len(m.s.Ledger)-ledgerMax:]
	}
}

func (m *Machine) recordWindfall(label string, tokens float64) {
	if tokens == 0 {
		return
	}
	m.record(Transaction{Kind: TxWindfall, Label: label, Tokens: tokens})
}
