// Package clean normalizes text copied from AI-CLI terminal output:
// strips terminal chrome (borders, padding), trims whitespace, and
// optionally rejoins lines that the terminal hard-wrapped.
package clean

import "strings"

// Opts controls the cleanup pipeline.
type Opts struct {
	// StripANSI removes ANSI / OSC escape sequences. Off by default
	// because most terminals already strip these on copy; turning it
	// on keeps surviving codes from leaking through.
	StripANSI bool

	// NoRejoin disables the wrapped-line rejoin heuristic. Useful when
	// pasting pure code where any reflow is unwanted.
	NoRejoin bool
}

// Clean runs the full cleanup pipeline on text and returns the result.
// Order is fixed; see the package doc and the project plan for rationale.
func Clean(text string, opts Opts) string {
	if opts.StripANSI {
		text = stripANSI(text)
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	lines = stripLeadingChrome(lines)
	lines = stripTrailingChrome(lines)

	if !opts.NoRejoin {
		lines = rejoinWrapped(lines)
	}

	lines = collapseBlankRuns(lines)

	return strings.Join(lines, "\n")
}

// collapseBlankRuns reduces any run of 3+ blank lines down to 2.
// Long blank runs usually come from bordered output where every "blank"
// row was full of padding; once stripped, they collapse into many empties.
func collapseBlankRuns(lines []string) []string {
	out := make([]string, 0, len(lines))
	blanks := 0
	for _, l := range lines {
		if l == "" {
			blanks++
			if blanks <= 2 {
				out = append(out, l)
			}
			continue
		}
		blanks = 0
		out = append(out, l)
	}
	return out
}
