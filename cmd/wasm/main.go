//go:build js && wasm

// Package main is the WASM entry point for the in-browser ai-clean demo.
// It exposes a single function `aiClean(text, opts) -> {text, stats}` to
// the host page, then blocks forever so the runtime stays alive.
package main

import (
	"syscall/js"

	"github.com/TheAndruu/ai-clean/internal/clean"
)

func main() {
	js.Global().Set("aiClean", js.FuncOf(aiClean))
	<-make(chan struct{})
}

func aiClean(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return errorResult("aiClean: missing text argument")
	}
	text := args[0].String()

	var opts clean.Opts
	if len(args) >= 2 && args[1].Type() == js.TypeObject {
		opts.StripANSI = optBool(args[1], "stripANSI")
		opts.NoRejoin = optBool(args[1], "noRejoin")
	}

	out, stats := clean.Clean(text, opts)
	return map[string]any{
		"text":  out,
		"stats": statsToJS(stats),
	}
}

func optBool(v js.Value, key string) bool {
	f := v.Get(key)
	if f.Type() != js.TypeBoolean {
		return false
	}
	return f.Bool()
}

func errorResult(msg string) map[string]any {
	return map[string]any{"text": "", "stats": map[string]any{"error": msg}}
}

func statsToJS(s clean.Stats) map[string]any {
	return map[string]any{
		"leadingBorderChar":     runeToString(s.LeadingBorderChar),
		"leadingBorderLines":    s.LeadingBorderLines,
		"trailingBorderChar":    runeToString(s.TrailingBorderChar),
		"trailingBorderLines":   s.TrailingBorderLines,
		"dedentColumns":         s.DedentColumns,
		"boxBorderLinesRemoved": s.BoxBorderLinesRemoved,
		"rejoinedLines":         s.RejoinedLines,
		"blankRunsCollapsed":    s.BlankRunsCollapsed,
		"leadingCapHit":         s.LeadingCapHit,
		"trailingCapHit":        s.TrailingCapHit,
		"unclosedFence":         s.UnclosedFence,
		"markdownTableSkipped":  s.MarkdownTableSkipped,
	}
}

func runeToString(r rune) string {
	if r == 0 {
		return ""
	}
	return string(r)
}
