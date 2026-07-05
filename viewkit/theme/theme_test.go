package theme

import "testing"

func TestNamedReturnsRegisteredTheme(t *testing.T) {
	th, ok := Named("monokai")
	if !ok {
		t.Fatal("Named(monokai) not found")
	}
	if got := th.Accent.GetForeground(); got != monokaiPalette.Accent {
		t.Fatalf("monokai accent = %v, want %v", got, monokaiPalette.Accent)
	}

	if _, ok := Named("does-not-exist"); ok {
		t.Fatal("Named(unknown) should report not found")
	}
}

func TestUseSyncsExportedVars(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	th, _ := Named("solarized-dark")
	Use(th)

	if Cur().Accent.GetForeground() != solarizedDarkPalette.Accent {
		t.Fatal("Cur() not updated by Use()")
	}
	if AccentSty.GetForeground() != solarizedDarkPalette.Accent {
		t.Fatalf("AccentSty not synced: %v", AccentSty.GetForeground())
	}
	if DimSty.GetForeground() != solarizedDarkPalette.Muted {
		t.Fatalf("DimSty not synced: %v", DimSty.GetForeground())
	}
}

func TestKeysDefaultFirst(t *testing.T) {
	keys := Keys()
	if len(keys) == 0 || keys[0] != "default" {
		t.Fatalf("Keys() = %v, want default first", keys)
	}
}
