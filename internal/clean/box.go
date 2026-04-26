package clean

// stripFullBoxBorders removes lines that are pure box-drawing chrome:
// horizontal rules and corners that frame a fully-boxed CLI panel
// (e.g. ┌─────┐ on top, └─────┘ on bottom). A line qualifies when
// every non-whitespace rune is in the Unicode box-drawing block
// (U+2500..U+257F) AND at least one rune is a horizontal-rule glyph.
// Markdown ASCII rules (---, ***, ===) are left alone — those are
// real content, not chrome.
func stripFullBoxBorders(lines []string, stats *Stats) []string {
	out := make([]string, 0, len(lines))
	removed := 0
	for _, l := range lines {
		if isFullBoxBorderLine(l) {
			removed++
			continue
		}
		out = append(out, l)
	}
	if stats != nil {
		stats.BoxBorderLinesRemoved += removed
	}
	return out
}

func isFullBoxBorderLine(l string) bool {
	hasContent := false
	hasHorizontal := false
	for _, r := range l {
		if r == ' ' || r == '\t' {
			continue
		}
		hasContent = true
		if r < 0x2500 || r > 0x257F {
			return false
		}
		if isHorizontalBoxRune(r) {
			hasHorizontal = true
		}
	}
	return hasContent && hasHorizontal
}

func isHorizontalBoxRune(r rune) bool {
	switch r {
	case '─', '━', '═', '╌', '╍', '╴', '╶', '╸', '╺', '╼', '╾':
		return true
	}
	return false
}
