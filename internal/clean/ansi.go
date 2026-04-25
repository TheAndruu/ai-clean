package clean

import "regexp"

var (
	// CSI: ESC [ ... letter — colors, cursor moves, mode sets.
	ansiCSI = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	// OSC: ESC ] ... terminator (BEL or ESC \) — titles, hyperlinks.
	ansiOSC = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
)

func stripANSI(s string) string {
	s = ansiCSI.ReplaceAllString(s, "")
	s = ansiOSC.ReplaceAllString(s, "")
	return s
}
