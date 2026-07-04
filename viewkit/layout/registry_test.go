package layout

import (
	"reflect"
	"testing"
)

func TestNewRegistryHasBuiltinLayouts(t *testing.T) {
	r := NewRegistry[testCtx]()
	got := r.LayoutKeys()
	want := []string{"flex", "grid", "single"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LayoutKeys = %v, want %v", got, want)
	}
}

func TestBuiltinFlexReadsParams(t *testing.T) {
	r := NewRegistry[testCtx]()
	l, err := r.layouts["flex"](Params{"minWidth": 25, "maxCols": 4})
	if err != nil {
		t.Fatalf("flex factory: %v", err)
	}
	fg, ok := l.(FlexGrid)
	if !ok {
		t.Fatalf("want FlexGrid, got %T", l)
	}
	if fg.MinWidth != 25 || fg.MaxCols != 4 {
		t.Fatalf("flex params = %+v, want {25 4}", fg)
	}
}

func TestBuiltinFlexDefaults(t *testing.T) {
	r := NewRegistry[testCtx]()
	l, _ := r.layouts["flex"](nil)
	fg := l.(FlexGrid)
	if fg.MinWidth != DefaultFlexMinWidth || fg.MaxCols != DefaultFlexMaxCols {
		t.Fatalf("flex defaults = %+v, want {%d %d}", fg, DefaultFlexMinWidth, DefaultFlexMaxCols)
	}
}

func TestPaneKeysStableOrder(t *testing.T) {
	r := NewRegistry[testCtx]()
	r.Pane("z", "Zed", func(testCtx) (Pane, bool) { return Pane{}, true })
	r.Pane("a", "Ay", func(testCtx) (Pane, bool) { return Pane{}, true })
	r.Pane("m", "Em", func(testCtx) (Pane, bool) { return Pane{}, true })

	keys := r.PaneKeys()
	want := []string{"z", "a", "m"}
	if len(keys) != 3 {
		t.Fatalf("PaneKeys len = %d, want 3", len(keys))
	}
	for i, k := range want {
		if keys[i].Key != k {
			t.Fatalf("PaneKeys[%d] = %q, want %q", i, keys[i].Key, k)
		}
	}
	if keys[0].Title != "Zed" {
		t.Fatalf("PaneInfo.Title = %q, want Zed", keys[0].Title)
	}
}

func TestPaneReRegisterKeepsOrder(t *testing.T) {
	r := NewRegistry[testCtx]()
	r.Pane("a", "A", func(testCtx) (Pane, bool) { return Pane{}, true })
	r.Pane("b", "B", func(testCtx) (Pane, bool) { return Pane{}, true })
	r.Pane("a", "A2", func(testCtx) (Pane, bool) { return Pane{}, true })

	keys := r.PaneKeys()
	if len(keys) != 2 || keys[0].Key != "a" || keys[1].Key != "b" {
		t.Fatalf("re-register changed order/count: %+v", keys)
	}
	if keys[0].Title != "A2" {
		t.Fatalf("re-register should update title, got %q", keys[0].Title)
	}
}

func TestBuildScreenNilRegistry(t *testing.T) {
	if _, err := BuildScreen[testCtx](ScreenSpec{Layout: "single"}, testCtx{}, nil); err == nil {
		t.Fatalf("nil registry should error")
	}
}
