package layout

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/theme"
)

func TestThemeInjectionFlowsThroughPanels(t *testing.T) {
	t.Cleanup(func() { theme.Use(theme.Default()) })

	custom := theme.Default()
	custom.TooNarrowTitle = "SCREEN TOO SMALL"
	theme.Use(custom)

	out := stripANSI(TooNarrow(theme.MinScreenWidth - 1))
	if !strings.Contains(out, "SCREEN TOO SMALL") {
		t.Fatalf("injected copy not rendered by TooNarrow: %q", out)
	}
	if strings.Contains(out, theme.DefaultTooNarrowTitle) {
		t.Fatalf("default copy leaked after Use: %q", out)
	}
}

func TestThemeUseIsRestorable(t *testing.T) {
	t.Cleanup(func() { theme.Use(theme.Default()) })

	theme.Use(theme.Theme{TooNarrowTitle: "X"})
	if got := theme.Cur().TooNarrowTitle; got != "X" {
		t.Fatalf("Cur did not reflect Use: %q", got)
	}

	theme.Use(theme.Default())
	if got := theme.Cur().TooNarrowTitle; got != theme.DefaultTooNarrowTitle {
		t.Fatalf("default not restored: %q", got)
	}
}
