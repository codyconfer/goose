package layout

import "github.com/codyconfer/viewkit/theme"

type Tier int

const (
	TierShort Tier = iota
	TierMedium
	TierTall
)

type Section struct {
	Content string
	MinTier Tier
}

func StackFit(tier Tier, sections ...Section) string {
	contents := make([]string, 0, len(sections))
	for _, s := range sections {
		if tier >= s.MinTier {
			contents = append(contents, s.Content)
		}
	}
	return Stack(contents...)
}

func BodyBudget(height int) int {
	if height <= 0 {
		return theme.MinBodyHeight - theme.AppMarginY*2
	}
	rows := height - theme.AppMarginY*2
	if rows < 1 {
		rows = 1
	}
	return rows
}

func ContentRows(height int) int {
	if height <= 0 {
		return 0
	}
	rows := height - theme.AppMarginY*2
	if rows < 1 {
		rows = 1
	}
	return rows
}

func TierForHeight(height int) Tier {
	rows := BodyBudget(height)
	switch {
	case rows >= theme.TallBodyHeight-theme.AppMarginY*2:
		return TierTall
	case rows >= theme.MinBodyHeight-theme.AppMarginY*2:
		return TierMedium
	default:
		return TierShort
	}
}

type TierRows struct{ Short, Medium, Tall int }

func (r TierRows) At(t Tier) int {
	switch t {
	case TierTall:
		return r.Tall
	case TierMedium:
		return r.Medium
	default:
		return r.Short
	}
}
