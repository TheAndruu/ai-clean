package clean

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// rejoinMinDocWidth: the document's longest line must be at least this
// many runes for rejoining to be considered. Below this, "near terminal
// width" can't be reliably inferred and we leave the text alone.
const rejoinMinDocWidth = 40

// rejoinWrapped merges lines that look like they were hard-wrapped by
// the terminal. Conservative by design: when in doubt, leave a newline.
//
// Skipped contexts:
//   - documents whose longest line is below rejoinMinDocWidth runes
//   - inside fenced code blocks (``` toggles)
//   - either side of the join has any leading whitespace (indented code,
//     list continuations, structured layout)
//   - list items (-, *, +, "1. ", "> ")
//   - around blank lines (paragraph boundaries)
//
// For two adjacent prose lines A and B, A is joined with B (A + " " + B)
// only if all hold:
//   - A does not end with sentence-terminating punctuation (. ! ? : ;)
//   - B does not start with a capital letter, list marker, or heading marker (#)
//   - A is "near terminal width" (within 10 chars of the doc's max line
//     length), OR B carries 1–3 leading spaces of typographic residue
//     (a continuation signal left over from imperfect dedent upstream).
//
// The terminal-width proxy avoids reflowing intentional short lines; the
// residue alternative catches snippets pulled from a wider context where
// A is just the tail of a wrapped line we don't have visibility into.
func rejoinWrapped(lines []string) []string {
	if len(lines) < 2 {
		return lines
	}

	maxLen := 0
	for _, l := range lines {
		if n := utf8.RuneCountInString(l); n > maxLen {
			maxLen = n
		}
	}
	if maxLen < rejoinMinDocWidth {
		return lines
	}
	wrapBand := maxLen - 10

	out := make([]string, 0, len(lines))
	inFence := false

	for _, l := range lines {
		if isFenceMarker(l) {
			inFence = !inFence
			out = append(out, l)
			continue
		}

		if len(out) == 0 {
			out = append(out, l)
			continue
		}

		prev := out[len(out)-1]

		if !canRejoin(prev, l, inFence, wrapBand) {
			out = append(out, l)
			continue
		}

		out[len(out)-1] = prev + " " + strings.TrimLeft(l, " \t")
	}

	return out
}

var (
	listMarker = regexp.MustCompile(`^\s*([-*+]|\d+\.|>)\s`)
	headingRE  = regexp.MustCompile(`^#{1,6}\s`)
)

func isFenceMarker(l string) bool {
	t := strings.TrimSpace(l)
	return strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
}

func canRejoin(prev, cur string, inFence bool, wrapBand int) bool {
	if inFence {
		return false
	}
	if prev == "" || cur == "" {
		return false
	}
	// Any leading whitespace on either side means structure (indented
	// code, list continuation, layout) — never reflow into or out of it.
	if hasLeadingWS(prev) || hasLeadingWS(cur) {
		return false
	}
	// A new list item on cur is a sibling, not a continuation. prev being
	// a list item is fine — wrapped continuations of list items are normal.
	if listMarker.MatchString(cur) {
		return false
	}
	if headingRE.MatchString(cur) {
		return false
	}
	// Don't rejoin if the previous line ends a sentence.
	if endsSentence(prev) {
		return false
	}
	// Don't rejoin if the current line starts with a capital — likely
	// a new sentence/heading even without an explicit period above.
	first := firstNonSpaceRune(cur)
	if first != 0 && unicode.IsUpper(first) {
		return false
	}
	// Two independent wrap signals: prev near terminal width, or cur with
	// 1–3 leading spaces of typographic residue. Either is enough.
	if !hasResidueLeadingWS(cur) && utf8.RuneCountInString(prev) < wrapBand {
		return false
	}
	return true
}

// hasLeadingWS treats only structural indentation as a rejoin block: a tab,
// or ≥4 leading spaces (markdown's indented-code-block boundary). Smaller
// leading runs are typographic residue from imperfect dedent.
func hasLeadingWS(l string) bool {
	if l == "" {
		return false
	}
	if l[0] == '\t' {
		return true
	}
	for i := 0; i < 4 && i < len(l); i++ {
		if l[i] != ' ' {
			return false
		}
	}
	return len(l) >= 4 && l[0] == ' '
}

// hasResidueLeadingWS: 1–3 leading spaces. Distinct from hasLeadingWS,
// which guards on tabs / 4+ spaces (markdown indented-code boundary).
func hasResidueLeadingWS(l string) bool {
	if l == "" || l[0] != ' ' {
		return false
	}
	n := 0
	for n < len(l) && l[n] == ' ' {
		n++
	}
	return n >= 1 && n <= 3
}

func endsSentence(l string) bool {
	if l == "" {
		return false
	}
	last := l[len(l)-1]
	return last == '.' || last == '!' || last == '?' || last == ':' || last == ';'
}

func firstNonSpaceRune(l string) rune {
	for _, r := range l {
		if r == ' ' || r == '\t' {
			continue
		}
		return r
	}
	return 0
}
