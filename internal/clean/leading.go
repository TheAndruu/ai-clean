package clean

import "strings"

// borderChars are the candidate left/right border characters used by
// CLIs to draw boxed output. Order doesn't matter; each is checked
// independently against the threshold.
var borderChars = []rune{'│', '┃', '|', '>', '┆', '╎', '┊', '┇'}

// borderThreshold: a candidate is accepted as a uniform border when
// it appears at the relevant position on at least this fraction of
// non-empty lines. 0.8 tolerates the occasional missing-border line
// while still rejecting characters that just happen to appear
// frequently in normal text.
const borderThreshold = 0.8

// stripLeadingChrome removes uniform leading whitespace and uniform
// leading border characters in alternating passes until a full pass
// produces no change. Capped at 3 iterations as a safety bound.
//
// Whitespace-dedent runs first within each pass: the minimum leading
// whitespace count across non-empty lines is computed and stripped from
// every non-empty line, preserving relative indentation of code blocks.
func stripLeadingChrome(lines []string) []string {
	for i := 0; i < 3; i++ {
		before := joinForCompare(lines)
		lines = dedentLeadingWhitespace(lines)
		lines = stripLeadingBorderChar(lines)
		if joinForCompare(lines) == before {
			break
		}
	}
	return lines
}

func dedentLeadingWhitespace(lines []string) []string {
	min := -1
	for _, l := range lines {
		if l == "" {
			continue
		}
		n := 0
		for _, r := range l {
			if r == ' ' || r == '\t' {
				n++
				continue
			}
			break
		}
		if n == len(l) {
			// All-whitespace line: treat as blank so it doesn't anchor min to 0.
			continue
		}
		if min == -1 || n < min {
			min = n
		}
		if min == 0 {
			break
		}
	}
	if min <= 0 {
		return lines
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if l == "" {
			out[i] = l
			continue
		}
		// Defensively cap at len(l) for short all-whitespace lines.
		cut := min
		if cut > len(l) {
			cut = len(l)
		}
		// Only cut if the leading chars are actually whitespace.
		safe := true
		for j := 0; j < cut; j++ {
			if l[j] != ' ' && l[j] != '\t' {
				safe = false
				break
			}
		}
		if safe {
			out[i] = l[cut:]
		} else {
			out[i] = l
		}
	}
	return out
}

func stripLeadingBorderChar(lines []string) []string {
	nonEmpty := 0
	for _, l := range lines {
		if l != "" {
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
			if l == "" {
				continue
			}
			rs := []rune(l)
			if len(rs) > 0 && rs[0] == ch {
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
		if l == "" {
			out[i] = l
			continue
		}
		rs := []rune(l)
		if len(rs) == 0 || rs[0] != best.ch {
			out[i] = l
			continue
		}
		// Drop the border char and one optional following space.
		rs = rs[1:]
		if len(rs) > 0 && rs[0] == ' ' {
			rs = rs[1:]
		}
		out[i] = string(rs)
	}
	return out
}

func joinForCompare(lines []string) string {
	return strings.Join(lines, "\n")
}
