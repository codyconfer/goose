package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func bgActiveEverywhere(s string) (bad rune, ok bool) {
	active := false
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && runes[j] != 'm' {
				j++
			}
			active = applySGR(string(runes[i+2:j]), active)
			i = j
			continue
		}
		if r == '\n' || r == ' ' {
			continue
		}
		if !active {
			return r, false
		}
	}
	return 0, true
}

func applySGR(params string, active bool) bool {
	toks := strings.Split(params, ";")
	for i := 0; i < len(toks); i++ {
		switch toks[i] {
		case "", "0":
			active = false
		case "49":
			active = false
		case "48":
			active = true
			if i+1 < len(toks) && toks[i+1] == "2" {
				i += 4
			} else if i+1 < len(toks) && toks[i+1] == "5" {
				i += 2
			}
		case "38":
			if i+1 < len(toks) && toks[i+1] == "2" {
				i += 4
			} else if i+1 < len(toks) && toks[i+1] == "5" {
				i += 2
			}
		}
	}
	return active
}

func TestScreenPaintsEveryCell(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)
	orig := *Cur()
	defer Use(orig)

	th, _ := Named("solarized-light")
	Use(th)

	line := "AAA" + ValSty.Render("BBB") + "   " + AccentSty.Render("CCC") + "  plain tail"
	body := AppFrame.Render(line + "\n" + DimSty.Render("second") + "   gap")
	out := Screen(body, 40, 8)

	if bad, ok := bgActiveEverywhere(out); !ok {
		t.Fatalf("found unpainted cell %q in:\n%q", string(bad), out)
	}
}

func TestScreenNoBgReturnsBodyUnchanged(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	Use(New(Palette{Text: lipgloss.Color("#ffffff")}))
	body := "hello\nworld"
	if got := Screen(body, 40, 8); got != body {
		t.Fatalf("Screen with empty Bg = %q, want unchanged", got)
	}
}
