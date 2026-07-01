package economy

import "github.com/codyconfer/goose/internal/content"

type Settings struct {
	LevelPace  int `json:"level_pace,omitempty"`
	EventPace  int `json:"event_pace,omitempty"`
	MarketPace int `json:"market_pace,omitempty"`
}

func DefaultSettings() Settings { return Settings{} }

func resolveIdx(spec content.SettingSpec, stored int) int {
	idx := stored - 1
	if idx < 0 || idx >= len(spec.Options) {
		idx = spec.Default
	}
	if idx < 0 || idx >= len(spec.Options) {
		idx = 0
	}
	return idx
}

func optionMult(spec content.SettingSpec, stored int) float64 {
	idx := resolveIdx(spec, stored)
	if idx < 0 || idx >= len(spec.Options) {
		return 1
	}
	return spec.Options[idx].Mult
}

func (s Settings) LevelMult() float64 { return optionMult(content.Settings.LevelPace, s.LevelPace) }

func (s Settings) EventMult() float64 { return optionMult(content.Settings.EventPace, s.EventPace) }

func (s Settings) MarketMult() float64 { return optionMult(content.Settings.MarketPace, s.MarketPace) }

func (s Settings) LevelIdx() int  { return resolveIdx(content.Settings.LevelPace, s.LevelPace) }
func (s Settings) EventIdx() int  { return resolveIdx(content.Settings.EventPace, s.EventPace) }
func (s Settings) MarketIdx() int { return resolveIdx(content.Settings.MarketPace, s.MarketPace) }

func (s Settings) WithLevel(idx int) Settings  { s.LevelPace = idx + 1; return s }
func (s Settings) WithEvent(idx int) Settings  { s.EventPace = idx + 1; return s }
func (s Settings) WithMarket(idx int) Settings { s.MarketPace = idx + 1; return s }

func (m *Machine) Settings() Settings { return m.s.Settings }

func (m *Machine) SetSettings(set Settings) { m.s.Settings = set }
