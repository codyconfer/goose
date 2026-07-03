package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/theme"
)

func ProgressBar(frac float64, width int) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac * float64(width))
	return theme.Cur().Accent.Render(strings.Repeat("█", filled)) + theme.Cur().Dim.Render(strings.Repeat("░", width-filled))
}

func Meter(frac float64, width int) string {
	return "[" + ProgressBar(frac, width) + "]"
}

func MeterWidth(frameWidth, desired int) int {
	if desired < 1 {
		return 1
	}
	max := frameWidth / 3
	if max < 8 {
		max = 8
	}
	if desired > max {
		return max
	}
	return desired
}

func Flash(message string) string {
	if message == "" {
		return ""
	}
	return theme.Cur().Dim.Italic(true).Render(message)
}

func Toggle(left, right string, leftActive bool) string {
	leftSty, rightSty := theme.Cur().Val, theme.Cur().Val
	if leftActive {
		leftSty = theme.Cur().Accent
	} else {
		rightSty = theme.Cur().Accent
	}
	return leftSty.Render(left) + theme.Cur().Dim.Render("  /  ") + rightSty.Render(right)
}

func ClampIndex(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= total {
		return total - 1
	}
	return index
}

func MoveIndex(index, delta, total int) int {
	return ClampIndex(index+delta, total)
}

func StepIndex(index, delta, total int) int {
	return MoveIndex(index, delta, total)
}
