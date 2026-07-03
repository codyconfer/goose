package keys

import "testing"

func TestMapResolvesAliases(t *testing.T) {
	m := NewMap(
		Binding{Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k"},
		Binding{Keys: []string{"down", "j"}, Action: Down},
	)
	for _, in := range []string{"up", "k"} {
		if a, ok := m.Action(in); !ok || a != Up {
			t.Fatalf("Action(%q) = %q,%v; want %q,true", in, a, ok, Up)
		}
	}
	if a, ok := m.Action("j"); !ok || a != Down {
		t.Fatalf("Action(j) = %q,%v; want %q,true", a, ok, Down)
	}
	if _, ok := m.Action("z"); ok {
		t.Fatal("Action(z) resolved; want unbound")
	}
}

func TestMatcherOnlyBindingEmitsNoHint(t *testing.T) {
	m := NewMap(
		Binding{Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k", Label: "select"},
		Binding{Keys: []string{"down", "j"}, Action: Down},
	)
	hints := m.Hints(Up, Down)
	if len(hints) != 1 {
		t.Fatalf("Hints len = %d; want 1 (Down is matcher-only)", len(hints))
	}
	if hints[0] != [2]string{"↑/↓/j/k", "select"} {
		t.Fatalf("Hints[0] = %v; want [↑/↓/j/k select]", hints[0])
	}
	if !m.Has(Down) {
		t.Fatal("Down should still be bound for matching")
	}
}

func TestDisplayGlyphFallsBackToKeys(t *testing.T) {
	b := Binding{Keys: []string{"y", "z"}, Action: Confirm}
	if got := b.DisplayGlyph(); got != "y/z" {
		t.Fatalf("DisplayGlyph = %q; want y/z", got)
	}
}

func TestHintLabeledOverridesLabel(t *testing.T) {
	m := NewMap(Binding{Keys: []string{"up"}, Action: Up, Glyph: "↑/↓/j/k", Label: "select"})
	if got := m.HintLabeled(Up, "scroll feed"); got != [2]string{"↑/↓/j/k", "scroll feed"} {
		t.Fatalf("HintLabeled = %v; want [↑/↓/j/k scroll feed]", got)
	}
}

func TestSchemeWithOverridesAreConfigurable(t *testing.T) {
	sc := Default().With(Binding{Keys: []string{"w"}, Action: Up, Glyph: "w"})
	m := NewMap(sc.Binding(Up))
	if a, ok := m.Action("w"); !ok || a != Up {
		t.Fatalf("Action(w) = %q,%v; want %q,true", a, ok, Up)
	}

	if _, ok := NewMap(Default().Binding(Up)).Action("w"); ok {
		t.Fatal("Default scheme mutated by With")
	}
}
