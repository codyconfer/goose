package economy

type PriceCandle struct {
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

func (m *Machine) PriceChart() ([]PriceCandle, int) {
	return append([]PriceCandle(nil), m.s.PriceCandles...), m.s.PriceCandleBeats
}

func (m *Machine) SetPriceChart(candles []PriceCandle, beats int) {
	if beats < 0 {
		beats = 0
	}
	m.s.PriceCandles = append([]PriceCandle(nil), candles...)
	m.s.PriceCandleBeats = beats
}
