package forms

import (
	"regexp"
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestConfirmSelectionAndResult(t *testing.T) {
	c := Confirm{Title: "DELETE", Message: "Remove save?"}
	if c.Yes {
		t.Fatal("zero value should default to No")
	}
	if got := c.Handle(keys.Left); got != Pending {
		t.Fatalf("Left = %v, want Pending", got)
	}
	if !c.Yes {
		t.Error("Left should select Yes")
	}
	if got := c.Handle(keys.Right); got != Pending || c.Yes {
		t.Error("Right should select No")
	}
	if got := c.Handle(keys.Confirm); got != Submitted {
		t.Errorf("Confirm = %v, want Submitted", got)
	}
	if got := c.Handle(keys.Cancel); got != Cancelled {
		t.Errorf("Cancel = %v, want Cancelled", got)
	}

	out := stripANSI(c.Render(layout.DefaultFrame()))
	for _, want := range []string{"DELETE", "Remove save?", "Yes", "No"} {
		if !strings.Contains(out, want) {
			t.Errorf("confirm render missing %q:\n%s", want, out)
		}
	}
}

func TestConfirmOverlay(t *testing.T) {
	bg := strings.TrimRight(strings.Repeat(strings.Repeat(".", 50)+"\n", 10), "\n")
	c := Confirm{Title: "OK?", Message: "sure"}
	out := stripANSI(c.Overlay(bg, layout.NewFrame(28)))
	if !strings.Contains(out, "OK?") || !strings.Contains(out, ".") {
		t.Fatalf("overlay should show prompt over background:\n%s", out)
	}
}

func TestFormTextInput(t *testing.T) {
	fm := NewForm(Field{Key: "name", Label: "Name", Kind: FieldText})
	fm.Insert("Jerry")
	fm.Insert("\n")
	if got := fm.Values()["name"]; got != "Jerry" {
		t.Fatalf("text value = %q, want Jerry", got)
	}
	fm.Handle(keys.Erase)
	if got := fm.Values()["name"]; got != "Jerr" {
		t.Errorf("after erase = %q, want Jerr", got)
	}
}

func TestFormMultilineKeepsNewlines(t *testing.T) {
	fm := NewForm(Field{Key: "bio", Label: "Bio", Kind: FieldMultiline})
	fm.Insert("line1\nline2")
	if got := fm.Values()["bio"]; got != "line1\nline2" {
		t.Fatalf("multiline value = %q", got)
	}
}

func TestFormSelectAndToggle(t *testing.T) {
	fm := NewForm(
		Field{Key: "risk", Label: "Risk", Kind: FieldSelect, Options: []string{"low", "med", "high"}},
		Field{Key: "auto", Label: "Auto", Kind: FieldToggle},
	)

	fm.Handle(keys.Right)
	fm.Handle(keys.Right)
	if got := fm.Values()["risk"]; got != "high" {
		t.Errorf("select value = %q, want high", got)
	}

	fm.Handle(keys.Right)
	if got := fm.Values()["risk"]; got != "high" {
		t.Errorf("select should clamp at high, got %q", got)
	}

	fm.Handle(keys.Down)
	if !fm.Handle(keys.Confirm) {
		t.Error("Confirm on toggle should be consumed")
	}
	if got := fm.Values()["auto"]; got != true {
		t.Errorf("toggle value = %v, want true", got)
	}
}

func TestFormMultiselectAndRadio(t *testing.T) {
	fm := NewForm(
		Field{Key: "tags", Label: "Tags", Kind: FieldMultiselect, Options: []string{"a", "b", "c"}},
		Field{Key: "tier", Label: "Tier", Kind: FieldRadio, Options: []string{"x", "y"}},
	)

	fm.Handle(keys.Confirm)
	fm.Handle(keys.Right)
	fm.Handle(keys.Right)
	fm.Handle(keys.Confirm)
	tags, ok := fm.Values()["tags"].([]string)
	if !ok || len(tags) != 2 || tags[0] != "a" || tags[1] != "c" {
		t.Fatalf("multiselect value = %v, want [a c]", fm.Values()["tags"])
	}

	fm.Handle(keys.Down)
	fm.Handle(keys.Right)
	if got := fm.Values()["tier"]; got != "y" {
		t.Errorf("radio value = %q, want y", got)
	}
}

func TestFormUnconsumedConfirmSignalsSubmit(t *testing.T) {
	fm := NewForm(Field{Key: "name", Label: "Name", Kind: FieldText})
	if fm.Handle(keys.Confirm) {
		t.Error("Confirm on text field should be unconsumed (host treats as submit)")
	}
}

func TestFormRenderShowsAllFields(t *testing.T) {
	fm := NewForm(
		Field{Key: "name", Label: "Name", Kind: FieldText, Text: "Wook"},
		Field{Key: "tags", Label: "Tags", Kind: FieldMultiselect, Options: []string{"a", "b"}},
	)
	out := stripANSI(fm.Render(layout.DefaultFrame(), "PROFILE"))
	for _, want := range []string{"PROFILE", "Name", "Wook", "Tags", "[ ] a", "[ ] b"} {
		if !strings.Contains(out, want) {
			t.Errorf("form render missing %q:\n%s", want, out)
		}
	}
}
