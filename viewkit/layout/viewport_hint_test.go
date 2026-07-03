package layout

import (
	"strings"
	"testing"
)

func TestViewportDirectionalHints(t *testing.T) {
	body := strings.Join([]string{"one", "two", "three", "four", "five"}, "\n")

	top := Viewport(body, 4, 0)
	if !strings.Contains(top, "▼") {
		t.Errorf("top of viewport should show a down arrow:\n%s", top)
	}
	if strings.Contains(top, "▲") {
		t.Errorf("top of viewport should not show an up arrow:\n%s", top)
	}

	mid := Viewport(body, 4, 1)
	if !strings.Contains(mid, "▲") || !strings.Contains(mid, "▼") {
		t.Errorf("middle of viewport should show both arrows:\n%s", mid)
	}

	bottom := Viewport(body, 4, 99)
	if !strings.Contains(bottom, "▲") {
		t.Errorf("bottom of viewport should show an up arrow:\n%s", bottom)
	}
	if strings.Contains(bottom, "▼") {
		t.Errorf("bottom of viewport should not show a down arrow:\n%s", bottom)
	}
}
