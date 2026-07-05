package panels

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestMarkdownStructure(t *testing.T) {
	src := strings.Join([]string{
		"# Title",
		"",
		"Some **bold** and *italic* and `code`.",
		"",
		"- one",
		"- two",
		"",
		"1. first",
		"2. second",
		"",
		"> a quote",
		"",
		"[link](http://x)",
		"---",
	}, "\n")

	out := stripANSI(Markdown(layout.DefaultFrame(), src))
	for _, want := range []string{
		"Title", "bold", "italic", "code",
		"• one", "• two",
		"1. first", "2. second",
		"┃ a quote",
		"link", "(http://x)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown output missing %q:\n%s", want, out)
		}
	}

	for _, bad := range []string{"**", "`code`", "](http"} {
		if strings.Contains(out, bad) {
			t.Errorf("markdown left raw marker %q:\n%s", bad, out)
		}
	}
}

func TestMarkdownFencedCodePreserved(t *testing.T) {
	src := "```\nx := 1\n```"
	out := stripANSI(Markdown(layout.DefaultFrame(), src))
	if !strings.Contains(out, "x := 1") {
		t.Fatalf("code block body missing:\n%s", out)
	}
	if strings.Contains(out, "```") {
		t.Errorf("fence markers should be stripped:\n%s", out)
	}
}

func TestMarkdownPanelHasTitle(t *testing.T) {
	out := stripANSI(MarkdownPanel(layout.DefaultFrame(), "DOCS", "hello"))
	for _, want := range []string{"DOCS", "hello"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestMarkdownWrapsLongParagraph(t *testing.T) {
	long := strings.Repeat("word ", 60)
	f := layout.NewFrame(30)
	out := stripANSI(Markdown(f, long))
	for _, line := range strings.Split(out, "\n") {
		if w := len([]rune(line)); w > f.BodyWidth() {
			t.Errorf("wrapped line width %d exceeds %d: %q", w, f.BodyWidth(), line)
		}
	}
}
