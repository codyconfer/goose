package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStackFitNoLimitMatchesStack(t *testing.T) {
	secs := []Section{
		{Content: "alpha", Priority: Essential},
		{Content: "bravo", Priority: 30},
		{Content: "charlie", Priority: 10},
	}
	got := StackFit(0, secs...)
	want := Stack("alpha", "bravo", "charlie")
	if got != want {
		t.Fatalf("StackFit(0) = %q, want Stack output %q", got, want)
	}
}

func TestStackFitEverythingFits(t *testing.T) {
	// Three single-line sections stack to 3 + 2 separators = 5 rows.
	secs := []Section{
		{Content: "alpha", Priority: Essential},
		{Content: "bravo", Priority: 30},
		{Content: "charlie", Priority: 10},
	}
	got := StackFit(5, secs...)
	if h := lipgloss.Height(got); h != 5 {
		t.Fatalf("height = %d, want 5:\n%s", h, got)
	}
	for _, want := range []string{"alpha", "bravo", "charlie"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to survive:\n%s", want, got)
		}
	}
}

func TestStackFitDropsHighestPriorityFirst(t *testing.T) {
	secs := []Section{
		{Content: "alpha", Priority: Essential},
		{Content: "bravo", Priority: 30},
		{Content: "charlie", Priority: 10},
	}
	// Budget 3 forces one drop: alpha + one survivor = 1 + 1 + 1 separator = 3.
	got := StackFit(3, secs...)
	if h := lipgloss.Height(got); h != 3 {
		t.Fatalf("height = %d, want 3:\n%s", h, got)
	}
	if strings.Contains(got, "bravo") {
		t.Errorf("bravo (priority 30) should drop before charlie (priority 10):\n%s", got)
	}
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "charlie") {
		t.Errorf("essential and lower-priority section should survive:\n%s", got)
	}
}

func TestStackFitTieDropsBottomMost(t *testing.T) {
	secs := []Section{
		{Content: "top", Priority: 20},
		{Content: "bottom", Priority: 20},
	}
	// Budget 1 leaves room for a single line; the bottom-most tie loses first,
	// then the remaining tie is dropped too since nothing is essential.
	got := StackFit(1, secs...)
	if h := lipgloss.Height(got); h != 1 {
		t.Fatalf("height = %d, want 1:\n%s", h, got)
	}

	// Budget 3 keeps both (2 + 1 separator); budget 2 forces exactly one drop
	// and it must be the bottom-most.
	got = StackFit(2, secs...)
	if strings.Contains(got, "bottom") || !strings.Contains(got, "top") {
		t.Errorf("bottom-most section should drop first on a tie:\n%s", got)
	}
}

func TestStackFitKeepsEssentialsWhenOverBudget(t *testing.T) {
	secs := []Section{
		{Content: "keep-1", Priority: Essential},
		{Content: "drop", Priority: 40},
		{Content: "keep-2", Priority: Essential},
	}
	// Budget 1 can't fit both essentials, but essentials are never dropped.
	got := StackFit(1, secs...)
	if strings.Contains(got, "drop") {
		t.Errorf("droppable section should be gone:\n%s", got)
	}
	if !strings.Contains(got, "keep-1") || !strings.Contains(got, "keep-2") {
		t.Errorf("essentials must survive even when over budget:\n%s", got)
	}
}

func TestStackFitSkipsEmptySections(t *testing.T) {
	// Empty sections contribute no height and no separator, matching Stack.
	got := StackFit(3, Section{Content: "alpha", Priority: Essential}, Section{Content: "", Priority: 10}, Section{Content: "bravo", Priority: Essential})
	want := Stack("alpha", "bravo")
	if got != want {
		t.Fatalf("StackFit with empty section = %q, want %q", got, want)
	}
}
