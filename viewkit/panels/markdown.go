package panels

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

var (
	mdBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalic  = regexp.MustCompile(`(?:\*([^*]+)\*|_([^_]+)_)`)
	mdCode    = regexp.MustCompile("`([^`]+)`")
	mdLink    = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	mdOrdered = regexp.MustCompile(`^(\d+)\. +(.*)$`)
)

func Markdown(f layout.Frame, src string) string {
	width := f.BodyWidth()
	t := theme.Cur()

	var out []string
	inCode := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimRight(raw, " ")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, t.Dim.Render(ansi.Truncate(line, width, "…")))
			continue
		}

		switch {
		case trimmed == "":
			out = append(out, "")
		case trimmed == "---" || trimmed == "***" || trimmed == "___":
			out = append(out, f.Rule())
		case strings.HasPrefix(trimmed, "### "):
			out = append(out, t.Accent.Render(ansi.Truncate(strings.TrimPrefix(trimmed, "### "), width, "…")))
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, t.Title.Render(ansi.Truncate(strings.TrimPrefix(trimmed, "## "), width, "…")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, t.Title.Bold(true).Render(ansi.Truncate(strings.TrimPrefix(trimmed, "# "), width, "…")))
		case strings.HasPrefix(trimmed, "> "):
			body := mdInline(strings.TrimPrefix(trimmed, "> "))
			out = append(out, wrapInline(t.Dim.Render("┃ "), body, width)...)
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			body := mdInline(trimmed[2:])
			out = append(out, wrapInline(t.Accent.Render("• "), body, width)...)
		case mdOrdered.MatchString(trimmed):
			mset := mdOrdered.FindStringSubmatch(trimmed)
			body := mdInline(mset[2])
			out = append(out, wrapInline(t.Accent.Render(mset[1]+". "), body, width)...)
		default:
			out = append(out, wrapInline("", mdInline(trimmed), width)...)
		}
	}
	return strings.Join(out, "\n")
}

func MarkdownPanel(f layout.Frame, title, src string) string {
	return f.Panel(title, strings.Split(Markdown(f, src), "\n")...)
}

func mdInline(s string) string {
	t := theme.Cur()
	s = mdCode.ReplaceAllStringFunc(s, func(m string) string {
		return t.Key.Render(mdCode.FindStringSubmatch(m)[1])
	})
	s = mdBold.ReplaceAllStringFunc(s, func(m string) string {
		return t.Accent.Bold(true).Render(mdBold.FindStringSubmatch(m)[1])
	})
	s = mdItalic.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdItalic.FindStringSubmatch(m)
		text := sub[1]
		if text == "" {
			text = sub[2]
		}
		return t.Val.Italic(true).Render(text)
	})
	s = mdLink.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdLink.FindStringSubmatch(m)
		return t.Accent.Render(sub[1]) + t.Dim.Render(" ("+sub[2]+")")
	})
	return s
}

func wrapInline(prefix, styled string, width int) []string {
	indent := strings.Repeat(" ", ansi.StringWidth(prefix))
	avail := width - ansi.StringWidth(prefix)
	if avail < 1 {
		avail = 1
	}
	wrapped := ansi.Wordwrap(styled, avail, "")
	lines := strings.Split(wrapped, "\n")
	for i, l := range lines {
		if i == 0 {
			lines[i] = prefix + l
		} else {
			lines[i] = indent + l
		}
	}
	return lines
}
