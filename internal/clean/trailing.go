package clean

import "strings"

// stripTrailingChrome removes uniform trailing border characters and
// trailing whitespace. Border-char detection runs before the whitespace
// trim because borders sit past padding (e.g., "text   │   "). The
// strip is looped (max 3 passes) to peel nested borders like "│ │".
func stripTrailingChrome(lines []string) []string {
	for i := 0; i < 3; i++ {
		before := joinForCompare(lines)
		lines = stripTrailingBorderChar(lines)
		for j, l := range lines {
			lines[j] = strings.TrimRight(l, " \t")
		}
		if joinForCompare(lines) == before {
			break
		}
	}
	return lines
}

func stripTrailingBorderChar(lines []string) []string {
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
	return out
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
