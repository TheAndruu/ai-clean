package clean

import "testing"

func FuzzClean(f *testing.F) {
	seeds := []string{
		"",
		"hello\nworld\n",
		"│ hello │\n│ world │\n│ done  │",
		"┌───┐\n│ a │\n└───┘",
		"this is a long line of prose that approaches terminal width here\n" +
			"continuation here.",
		"| col1 | col2 |\n| ---- | ---- |\n| a    | b    |\n| c    | d    |",
		"intro\n```\ncode without close",
		"\x1b[31mred\x1b[0m text",
		"  indented\n  more indented\n",
		"> > > deep nesting\n> > > of borders\n> > > here",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		out1, _ := Clean(s, Opts{})
		out2, _ := Clean(out1, Opts{})
		if out1 != out2 {
			t.Fatalf("Clean is not idempotent\ninput=%q\nfirst=%q\nsecond=%q", s, out1, out2)
		}
	})
}
