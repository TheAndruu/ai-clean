# CLAUDE.md

Project context for Claude Code sessions on this repo.

## What this is

`ai-clean` — a small Go CLI that reads text from the clipboard, normalizes it (strips terminal chrome, trims whitespace, optionally rejoins wrapped lines), and writes it back. The motivating use case is cleaning up output copied from AI CLI tools (Claude Code, Microsoft's Copilot CLI) that arrives with border characters like `│`, trailing whitespace padding, and terminal-hard-wrapped lines.

End-to-end workflow: **select → copy → `ai-clean` → paste**.

## Architecture

Single Go binary. Two packages:

- `main` (`./main.go`) — flag parsing (stdlib `flag`), clipboard I/O via `github.com/atotto/clipboard`, friendly error message when the Linux clipboard helper is missing.
- `internal/clean` — the cleanup pipeline. Public surface is `Clean(text string, opts Opts) string`.

## Cleanup pipeline (order matters)

1. `ansi.go` — optional ANSI/OSC strip. Off by default; only runs if `Opts.StripANSI` is true.
2. `leading.go` — `stripLeadingChrome`. Loops (max 3 passes) over: (a) `dedentLeadingWhitespace` strips the minimum-common leading whitespace across non-empty lines, and (b) `stripLeadingBorderChar` strips a uniform leading border char (one of `│ ┃ | > ┆ ╎ ┊ ┇`) when it appears on ≥80% of non-empty lines, plus an optional one trailing space. The loop handles nested borders (`│ │ text`).
3. `trailing.go` — `stripTrailingChrome`. Mirror of leading: loops over (a) `stripTrailingBorderChar` (looks past trailing whitespace to find the rightmost non-WS char) and (b) trailing whitespace trim. Border-char strip must run before whitespace trim — that's why it's bundled here, not in step 2.
4. `rejoin.go` — `rejoinWrapped`. Conservative wrapped-line merge. Skipped when `Opts.NoRejoin` is true. Guards: fenced-code detection, any leading whitespace on either side disables the merge, list/heading markers, sentence-terminator detection, and a minimum document-width floor (`rejoinMinDocWidth = 40`).
5. `clean.go` — `collapseBlankRuns`. Cosmetic: runs of 3+ blank lines collapse to 2.

## Conventions

- **Pure Go, no Cgo.** Keeps cross-compilation trivial (`GOOS=linux GOARCH=amd64 go build` works from anywhere). The clipboard library was chosen specifically to avoid Cgo. Don't introduce Cgo deps.
- **Tests are the correctness gate.** `internal/clean/clean_test.go` is table-driven plus a `testdata/`-driven full-pipeline test. New pipeline behavior should land with a test case. When fixing a bug, add the regression case to the table first.
- **The 80% border-detection threshold** (`borderThreshold` in `leading.go`) is intentional. Don't raise to 100% — that breaks on real-world output with one occasional missing border line.
- **Unicode care.** Border characters are multi-byte runes (`│` is 3 bytes). Use `[]rune` for indexing into lines, not `[]byte`. See `leading.go` and `trailing.go` for the pattern.
- **No comments explaining what code does.** Only the *why* — the package doc comments and the pipeline rationale. Test names are the documentation for behavior.

## Build / test

```sh
go test ./...                 # run all tests
go build -o ai-clean .         # build local binary
go install .                  # install to $GOBIN

# cross-compile
GOOS=linux   GOARCH=amd64 go build -o dist/ai-clean-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/ai-clean-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o dist/ai-clean-windows-amd64.exe .
```

## Releases

Distribution is via GoReleaser (`.goreleaser.yml`) and a GitHub Actions workflow triggered on `v*` tags. To cut a release: `git tag v0.X.0 && git push --tags`. CI builds and uploads binaries for darwin/linux/windows × amd64/arm64.

## Where things live

| File | Purpose |
|---|---|
| `main.go` | CLI flags, clipboard I/O, Linux helper detection |
| `internal/clean/clean.go` | `Clean()` orchestrator, `Opts`, blank-line collapse |
| `internal/clean/ansi.go` | Opt-in ANSI/OSC strip |
| `internal/clean/leading.go` | Leading whitespace dedent + leading border-char strip (looped) |
| `internal/clean/trailing.go` | Trailing border-char strip + trailing whitespace trim (looped) |
| `internal/clean/rejoin.go` | Wrapped-line rejoin heuristic |
| `internal/clean/clean_test.go` | Table-driven tests + testdata-driven full-pipeline tests |
| `testdata/*_sample.txt` | Real captured input |
| `testdata/*_expected.txt` | Hand-curated expected output for the matching sample |
