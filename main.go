package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"

	"github.com/TheAndruu/ai-clean/internal/clean"
)

var version = "dev"

func main() {
	var (
		stdinFlag   = flag.Bool("stdin", false, "read text from stdin and write cleaned text to stdout instead of using the clipboard")
		dryRun      = flag.Bool("dry-run", false, "print the cleaned text to stdout instead of writing it back to the clipboard")
		noRejoin    = flag.Bool("no-rejoin", false, "disable the wrapped-line rejoin heuristic (safer for pure code)")
		stripANSI   = flag.Bool("strip-ansi", false, "also strip ANSI / OSC escape sequences (off by default)")
		showVersion = flag.Bool("version", false, "print version and exit")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Cleans AI-CLI terminal output on the clipboard: strips border")
		fmt.Fprintln(os.Stderr, "characters, trailing whitespace, and rejoins terminal-wrapped lines.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	opts := clean.Opts{
		StripANSI: *stripANSI,
		NoRejoin:  *noRejoin,
	}

	if *stdinFlag && *dryRun {
		fmt.Fprintln(os.Stderr, "ai-clean: --dry-run applies to clipboard mode only; cannot combine with --stdin")
		os.Exit(2)
	}

	if *stdinFlag {
		runStdin(opts)
		return
	}

	runClipboard(opts, *dryRun)
}

func runStdin(opts clean.Opts) {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-clean: read stdin: %v\n", err)
		os.Exit(1)
	}
	out := clean.Clean(string(in), opts)
	if _, err := os.Stdout.WriteString(out); err != nil {
		fmt.Fprintf(os.Stderr, "ai-clean: write stdout: %v\n", err)
		os.Exit(1)
	}
}

func runClipboard(opts clean.Opts, dryRun bool) {
	if clipboard.Unsupported {
		fmt.Fprintln(os.Stderr, clipboardHelpMessage())
		os.Exit(1)
	}

	in, err := clipboard.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-clean: read clipboard: %v\n", err)
		if isLinuxClipboardMissing(err) {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, clipboardHelpMessage())
		}
		os.Exit(1)
	}

	out := clean.Clean(in, opts)

	if dryRun {
		fmt.Print(out)
		return
	}

	if err := clipboard.WriteAll(out); err != nil {
		fmt.Fprintf(os.Stderr, "ai-clean: write clipboard: %v\n", err)
		os.Exit(1)
	}

	lineCount := 0
	if out != "" {
		lineCount = strings.Count(out, "\n")
		if !strings.HasSuffix(out, "\n") {
			lineCount++
		}
	}
	fmt.Printf("✓ cleaned %d line(s)\n", lineCount)
}

func isLinuxClipboardMissing(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no clipboard utilities") ||
		strings.Contains(msg, "executable file not found") ||
		strings.Contains(msg, "xclip") ||
		strings.Contains(msg, "xsel")
}

func clipboardHelpMessage() string {
	return "ai-clean needs a clipboard helper.\n" +
		"  Linux:   install one of: xclip, xsel, or wl-clipboard\n" +
		"           e.g. `sudo apt install xclip` or `sudo pacman -S xclip`\n" +
		"  macOS:   pbcopy/pbpaste ship with the OS — no install needed\n" +
		"  Windows: clip.exe ships with the OS — no install needed\n" +
		"\n" +
		"You can also pipe text in via --stdin to bypass the clipboard."
}
