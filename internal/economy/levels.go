package economy

import "github.com/codyconfer/goose/internal/content"

var LevelThresholds = content.Levels

func MaxLevel() int { return len(LevelThresholds) }

func LevelFor(eggs float64) int { return levelFor(eggs, LevelThresholds) }

func levelFor(eggs float64, thresholds []float64) int {
	lvl := 1
	for i := 1; i < len(thresholds); i++ {
		if eggs >= thresholds[i] {
			lvl = i + 1
		}
	}
	return lvl
}

func NextLevelEggs(eggs float64) (float64, bool) {
	lvl := LevelFor(eggs)
	if lvl >= len(LevelThresholds) {
		return 0, false
	}
	return LevelThresholds[lvl], true
}

func (s State) thresholds() []float64 {
	mult := s.Settings.LevelMult()
	if mult == 1 {
		return LevelThresholds
	}
	out := make([]float64, len(LevelThresholds))
	for i, t := range LevelThresholds {
		out[i] = t * mult
	}
	return out
}

func (s State) LevelEggs() float64 {
	if s.PeakEggs > s.Eggs {
		return s.PeakEggs
	}
	return s.Eggs
}

func (s State) Level() int { return levelFor(s.LevelEggs(), s.thresholds()) }

func (s State) NextLevelEggs() (float64, bool) {
	thr := s.thresholds()
	lvl := s.Level()
	if lvl >= len(thr) {
		return 0, false
	}
	return thr[lvl], true
}

func (s State) LevelFloor() float64 {
	thr := s.thresholds()
	lvl := s.Level()
	if lvl-1 < 0 || lvl-1 >= len(thr) {
		return 0
	}
	return thr[lvl-1]
}
