package panels

import "strings"

func ViewportLayout(body string, rows, offset int) string {
	if rows <= 0 {
		return body
	}

	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return Viewport(body, rows, offset)
	}

	footerRows := CountLines(footer)
	if footerRows >= rows {
		return Viewport(footer, rows, 0)
	}

	contentRows := rows - footerRows
	separator := ""
	if content != "" && contentRows > 0 {
		contentRows--
		separator = "\n\n"
	}

	contentView := ""
	if contentRows > 0 && content != "" {
		contentView = Viewport(content, contentRows, offset)
		contentView = PadLines(contentView, contentRows)
	}

	switch {
	case contentView == "":
		return PadLines("", rows-footerRows) + footer
	case separator == "":
		return contentView + footer
	default:
		return contentView + separator + footer
	}
}

func ScrollableBody(body string, rows int) string {
	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return body
	}
	if ScrollableRows(body, rows) < 1 {
		return ""
	}
	return content
}

func ScrollableRows(body string, rows int) int {
	if rows <= 0 {
		return 0
	}

	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return rows
	}

	footerRows := CountLines(footer)
	if footerRows >= rows {
		return 0
	}

	contentRows := rows - footerRows
	if content != "" && contentRows > 0 {
		contentRows--
	}
	return max(contentRows, 0)
}

func SplitStickyFooter(body string) (content, footer string) {
	idx := strings.LastIndex(body, "\n\n")
	if idx < 0 {
		return body, ""
	}
	return body[:idx], body[idx+2:]
}

func CountLines(s string) int {
	lines := 0
	for range strings.SplitSeq(s, "\n") {
		lines++
	}
	if lines == 0 {
		return 1
	}
	return lines
}

func PadLines(body string, rows int) string {
	if rows <= 0 {
		return ""
	}
	if body == "" {
		return strings.Repeat("\n", max(rows-1, 0))
	}

	lines := CountLines(body)
	if lines >= rows {
		return body
	}

	var b strings.Builder
	if body != "" {
		b.WriteString(body)
	}
	for i := lines; i < rows; i++ {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
