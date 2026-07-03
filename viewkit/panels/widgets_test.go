package panels

import (
	"strings"
	"testing"
)

func TestMeterClampsToWidth(t *testing.T) {
	out := stripANSI(Meter(2, 4))
	if got := strings.Count(out, "█"); got != 4 {
		t.Fatalf("Meter filled cells = %d, want 4", got)
	}
}
