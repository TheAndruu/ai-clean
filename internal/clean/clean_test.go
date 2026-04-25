package clean

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClean(t *testing.T) {
	cases := []struct {
		name string
		in   string
		opts Opts
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "whitespace only collapses to single newline",
			in:   "   \n  \t  \n",
			want: "\n",
		},
		{
			name: "single line no newline",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "trailing whitespace stripped",
			in:   "hello   \nworld\t\t\n",
			want: "hello\nworld\n",
		},
		{
			name: "uniform leading whitespace dedented",
			in:   "  hello\n    code\n  text",
			want: "hello\n  code\ntext",
		},
		{
			name: "leading border char stripped",
			in:   "│ first\n│ second\n│ third",
			want: "first\nsecond\nthird",
		},
		{
			name: "leading pipe variant stripped",
			in:   "| one\n| two\n| three",
			want: "one\ntwo\nthree",
		},
		{
			name: "leading whitespace then border char",
			in:   "  │ a\n  │ b\n  │ c",
			want: "a\nb\nc",
		},
		{
			name: "fully bordered output",
			in:   "│ hello   │\n│ world   │\n│ done    │",
			want: "hello\nworld\ndone",
		},
		{
			name: "nested borders",
			in:   "│ │ a │ │\n│ │ b │ │\n│ │ c │ │",
			want: "a\nb\nc",
		},
		{
			name: "mid-line border char preserved",
			in:   "use the | pipe operator\nto chain | commands\nin your | shell",
			want: "use the | pipe operator\nto chain | commands\nin your | shell",
		},
		{
			name: "code block indentation preserved after dedent",
			in:   "  prose line\n      indented code\n      more code\n  back to prose",
			want: "prose line\n    indented code\n    more code\nback to prose",
		},
		{
			name: "blank lines preserved",
			in:   "│ first paragraph\n│\n│ second paragraph",
			want: "first paragraph\n\nsecond paragraph",
		},
		{
			name: "ANSI off by default leaves codes",
			in:   "\x1b[31mred\x1b[0m text",
			want: "\x1b[31mred\x1b[0m text",
		},
		{
			name: "ANSI on strips codes",
			in:   "\x1b[31mred\x1b[0m text",
			opts: Opts{StripANSI: true},
			want: "red text",
		},
		{
			name: "OSC hyperlink stripped when enabled",
			in:   "\x1b]8;;https://x.com\x07link\x1b]8;;\x07",
			opts: Opts{StripANSI: true},
			want: "link",
		},
		{
			name: "list items not rejoined",
			in:   "- first item\n- second item\n- third item",
			want: "- first item\n- second item\n- third item",
		},
		{
			name: "rejoin wrapped prose at terminal width",
			in: "this is a long line of prose that fills the terminal width\n" +
				"continuation of that same sentence here.\n" +
				"this is a long line of prose that fills the terminal width\n" +
				"continuation of that same sentence here.",
			want: "this is a long line of prose that fills the terminal width continuation of that same sentence here.\n" +
				"this is a long line of prose that fills the terminal width continuation of that same sentence here.",
		},
		{
			name: "no rejoin when disabled",
			in: "this is a long line of prose that fills the terminal width\n" +
				"continuation of that same sentence here.",
			opts: Opts{NoRejoin: true},
			want: "this is a long line of prose that fills the terminal width\n" +
				"continuation of that same sentence here.",
		},
		{
			name: "rejoin skips fenced code blocks",
			in:   "intro line\n```\nlong code line that almost reaches terminal width here\nshort\n```\noutro line",
			want: "intro line\n```\nlong code line that almost reaches terminal width here\nshort\n```\noutro line",
		},
		{
			name: "headings not rejoined into",
			in: "this line is fairly long and reaches near terminal width here\n" +
				"# Heading should not merge",
			want: "this line is fairly long and reaches near terminal width here\n" +
				"# Heading should not merge",
		},
		{
			name: "capital starting next line not rejoined",
			in: "this line is fairly long and reaches near terminal width here\n" +
				"Next sentence starts capitalized.",
			want: "this line is fairly long and reaches near terminal width here\n" +
				"Next sentence starts capitalized.",
		},
		{
			name: "triple+ blank runs collapsed",
			in:   "a\n\n\n\n\nb",
			want: "a\n\n\nb", // collapseBlankRuns keeps 2 blanks → 3 newlines between non-blanks
		},
		{
			name: "mixed prefix and non-prefix below threshold leaves all alone",
			in:   "│ a\nplain b\nplain c\nplain d\nplain e",
			want: "│ a\nplain b\nplain c\nplain d\nplain e",
		},
		{
			name: "prefix above threshold strips matching, leaves outliers",
			in:   "│ a\n│ b\n│ c\n│ d\nplain e",
			want: "a\nb\nc\nd\nplain e",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Clean(tc.in, tc.opts)
			if got != tc.want {
				t.Errorf("Clean mismatch\n--- input ---\n%q\n--- want ---\n%q\n--- got ---\n%q", tc.in, tc.want, got)
			}
		})
	}
}

func TestCleanFromTestdata(t *testing.T) {
	dir := "../../testdata"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("testdata dir not present: %v", err)
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, "_sample.txt") {
			continue
		}
		base := strings.TrimSuffix(name, "_sample.txt")
		expectedPath := filepath.Join(dir, base+"_expected.txt")
		t.Run(base, func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Skipf("no expected file %q", expectedPath)
				return
			}
			got := Clean(string(in), Opts{})
			if got != string(want) {
				t.Errorf("mismatch on %s\n--- want ---\n%s\n--- got ---\n%s", base, string(want), got)
			}
		})
	}
}
