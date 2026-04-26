package clean

import "strings"

// stripTrailingChrome removes uniform trailing border characters and
// trailing whitespace. Border-char detection runs before the whitespace
// trim because borders sit past padding (e.g., "text   │   "). The
// loop runs until convergence (each changing pass strictly shrinks the
// document). If more than nestingWarnThreshold passes were needed,
// stats.TrailingCapHit is set so --explain can surface it.
func stripTrailingChrome(lines []string, stats *Stats) []string {
	passes := 0
	for passes < pipelineSafetyCap {
		before := joinForCompare(lines)
		lines = stripTrailingBorderChar(lines, stats)
		for j, l := range lines {
			lines[j] = strings.TrimRight(l, " \t")
		}
		if joinForCompare(lines) == before {
			break
		}
		passes++
	}
	if stats != nil && passes > nestingWarnThreshold {
		stats.TrailingCapHit = true
	}
	return lines
}

func stripTrailingBorderChar(lines []string, stats *Stats) []string {
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		return lines
	}

	type cand struct {
		ch    rune
		count int
	}
	var best cand
	for _, ch := range borderChars {
		c := 0
		for _, l := range lines {
			if rightmostNonWS(l) == ch {
				c++
			}
		}
		if c > best.count {
			best = cand{ch: ch, count: c}
		}
	}

	if float64(best.count)/float64(nonEmpty) < borderThreshold {
		return lines
	}

	if best.ch == '|' && looksLikeMarkdownTableTrailing(lines) {
		if stats != nil {
			stats.MarkdownTableSkipped++
		}
		return lines
	}

	out := make([]string, len(lines))
	for i, l := range lines {
		if rightmostNonWS(l) != best.ch {
			out[i] = l
			continue
		}
		// Find index of the border char (last occurrence past trailing WS).
		trimmed := strings.TrimRight(l, " \t")
		rs := []rune(trimmed)
		// rs[len-1] is the border char (since rightmostNonWS matched).
		rs = rs[:len(rs)-1]
		// Also drop one space immediately preceding it, if present.
		if len(rs) > 0 && rs[len(rs)-1] == ' ' {
			rs = rs[:len(rs)-1]
		}
		out[i] = string(rs)
	}
	if stats != nil {
		if stats.TrailingBorderChar == 0 {
			stats.TrailingBorderChar = best.ch
		}
		stats.TrailingBorderLines += best.count
	}
	return out
}

// looksLikeMarkdownTableTrailing returns true when lines ending with '|'
// also have at least one interior '|' before the trailing one — i.e. the
// trailing '|' is a table cell-closer, not a border.
func looksLikeMarkdownTableTrailing(lines []string) bool {
	trailing := 0
	withInterior := 0
	for _, l := range lines {
		rs := []rune(strings.TrimRight(l, " \t"))
		if len(rs) == 0 || rs[len(rs)-1] != '|' {
			continue
		}
		trailing++
		// Skip a leading '|' if present (cell-opening pipe).
		start := 0
		if rs[0] == '|' {
			start = 1
		}
		for j := start; j < len(rs)-1; j++ {
			if rs[j] == '|' {
				withInterior++
				break
			}
		}
	}
	if trailing == 0 {
		return false
	}
	return float64(withInterior)/float64(trailing) >= markdownTableThreshold
}

// rightmostNonWS returns the last non-(space/tab) rune in a line, or 0
// if the line is empty or all whitespace.
func rightmostNonWS(l string) rune {
	rs := []rune(l)
	for i := len(rs) - 1; i >= 0; i-- {
		if rs[i] != ' ' && rs[i] != '\t' {
			return rs[i]
		}
	}
	return 0
}
