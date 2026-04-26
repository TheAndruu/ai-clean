# CLAUDE.md

Project context for Claude Code sessions on this repo.

## What this is

`ai-clean` — a small Go CLI that reads text from the clipboard, normalizes it (strips terminal chrome, trims whitespace, optionally rejoins wrapped lines), and writes it back. The motivating use case is cleaning up output copied from AI CLI tools (Claude Code, Microsoft's Copilot CLI) that arrives with border characters like `│`, trailing whitespace padding, and terminal-hard-wrapped lines.

End-to-end workflow: **select → copy → `ai-clean` → paste**.

## Architecture

Single Go binary. Two packages:

- `main` (`./main.go`) — flag parsing (stdlib `flag`), clipboard I/O via `github.com/atotto/clipboard`, friendly error message when the Linux clipboard helper is missing. The work is split into `run(args, stdin, stdout, stderr) int` so `main_test.go` can exercise the CLI without spawning a subprocess.
- `internal/clean` — the cleanup pipeline. Public surface is `Clean(text string, opts Opts) (string, Stats)`. `Stats` reports per-stage counts and warning flags; `main.go` consumes it for `--explain`.

## Cleanup pipeline (order matters)

The whole pipeline (steps 2–5) runs inside a fix-point loop in `clean.go` until a full pass makes no change. This is required for correctness: each stage can produce input another stage would clean further (a trailing-strip can turn a previously-mixed line into pure box chrome; rejoin can expose leading whitespace from the merged tail; etc.). Convergence is guaranteed because every changing pass strictly shrinks the document.

1. `ansi.go` — optional ANSI/OSC strip. Off by default; only runs if `Opts.StripANSI` is true. Runs once at input, before the loop.
2. `box.go` — `stripFullBoxBorders`. Removes lines that are pure box-drawing chrome: every non-WS rune is in `U+2500..U+257F` AND at least one is a horizontal rule (`─ ━ ═ ╌ ╍`). Catches `┌─┐`/`└─┘`/`═══` framing without touching markdown ASCII rules (`---`, `***`, `===`).
3. `leading.go` — `stripLeadingChrome`. Loops to convergence over: (a) `dedentLeadingWhitespace` strips the minimum-common leading whitespace across non-empty lines, and (b) `stripLeadingBorderChar` strips a uniform leading border char (one of `│ ┃ | > ┆ ╎ ┊ ┇`) when it appears on ≥80% of non-empty lines, plus an optional one trailing space. **Markdown-table guard**: if the candidate border is `|` and ≥50% of border-having lines also contain an interior `|`, the strip is skipped (`looksLikeMarkdownTable`). If more than `nestingWarnThreshold` (3) passes were needed, `Stats.LeadingCapHit` is set.
4. `trailing.go` — `stripTrailingChrome`. Mirror of leading: loops to convergence over (a) `stripTrailingBorderChar` (looks past trailing whitespace to find the rightmost non-WS char; same markdown-table guard) and (b) trailing whitespace trim. Border-char strip must run before whitespace trim — that's why it's bundled here, not in step 3.
5. `rejoin.go` — `rejoinWrapped`. Conservative wrapped-line merge. Skipped when `Opts.NoRejoin` is true. Guards: fenced-code detection (sets `Stats.UnclosedFence` if a fence opens without closing), any leading whitespace on either side disables the merge, list/heading markers, markdown table rows (`isTableRowLine`), sentence-terminator detection, and a minimum document-width floor (`rejoinMinDocWidth = 40`).
6. `clean.go` — `collapseBlankRuns`. Cosmetic: runs of 3+ blank lines collapse to 2. Runs once after the loop.

CRLF and lone-CR normalization happens at input, before the loop: `\r\n` → `\n`, then any remaining `\r` → `\n` (handles old-Mac line endings and stray CRs that would otherwise break idempotency).

## Conventions

- **Pure Go, no Cgo.** Keeps cross-compilation trivial (`GOOS=linux GOARCH=amd64 go build` works from anywhere). The clipboard library was chosen specifically to avoid Cgo. Don't introduce Cgo deps.
- **Tests are the correctness gate.** `internal/clean/clean_test.go` is table-driven plus a `testdata/`-driven full-pipeline test. New pipeline behavior should land with a test case. When fixing a bug, add the regression case to the table first.
- **Idempotency is an invariant.** `Clean(Clean(x), opts) == Clean(x, opts)` must hold for all inputs. Enforced by `TestCleanIdempotentOverTestdata` (every fixture) and `FuzzClean` (random inputs). If you touch the pipeline, run the fuzzer for at least 30 seconds — it has caught real bugs (residue from nested-border peeling, lone `\r`, mixed-content lines turning into pure box chrome).
- **Pipeline ordering is fragile.** Stages can produce input other stages would clean further. The fix-point loop in `Clean()` (`clean.go`) handles this. Don't move work outside the loop unless you're sure it can't introduce new chrome the other stages would catch.
- **The 80% border-detection threshold** (`borderThreshold` in `leading.go`) is intentional. Don't raise to 100% — that breaks on real-world output with one occasional missing border line.
- **The 50% markdown-table threshold** (`markdownTableThreshold`) is lower than the border threshold because table separators (`|---|---|`) and continuation rows can dilute the interior-`|` count.
- **The 3-pass `nestingWarnThreshold` is a warning, not an execution cap.** The loops run until convergence; the threshold only decides whether `Stats.LeadingCapHit` / `TrailingCapHit` are set. The hard cap (`pipelineSafetyCap = 100`) exists only to make a heuristic bug fail loudly instead of looping forever.
- **Unicode care.** Border characters are multi-byte runes (`│` is 3 bytes). Use `[]rune` for indexing into lines, not `[]byte`. See `leading.go` and `trailing.go` for the pattern.
- **No comments explaining what code does.** Only the *why* — the package doc comments and the pipeline rationale. Test names are the documentation for behavior.

## Build / test

```sh
go test ./...                                          # run all tests
go test ./internal/clean -fuzz=FuzzClean -fuzztime=30s # ad-hoc fuzz (idempotency)
go test ./internal/clean -bench=BenchmarkClean -benchmem -run=^$  # baseline perf
go build -o ai-clean .                                 # build local binary
go install .                                           # install to $GOBIN

# cross-compile
GOOS=linux   GOARCH=amd64 go build -o dist/ai-clean-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/ai-clean-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o dist/ai-clean-windows-amd64.exe .
```

## Adding example test cases

End-to-end fixtures live in `testdata/examples/` as `<characteristic>_sample.txt` / `<characteristic>_expected.txt` pairs. Naming describes what the case exercises (e.g. `wrapped_padded_indented`, `full_border_padded`), not which tool produced the output — the cleanup heuristics are source-agnostic.

Workflow for adding a new example:

1. Save the raw captured output as `testdata/examples/<name>_sample.txt`.
2. Run `go test ./internal/clean -run TestCleanFromTestdata -update` to generate a candidate `<name>_expected.txt` from the current `Clean()` output.
3. Review with `git diff testdata/examples/`. If the output is correct, commit both files.
4. If the output is wrong, the case has revealed a rule gap. Edit the relevant `internal/clean/*.go` file, re-run with `-update`, review the diff again. Repeat until the expected output is right, then commit.

Without `-update`, a missing or mismatched expected file fails the test — intentional, so silent drift can't sneak in.

## Releases

Distribution is via GoReleaser (`.goreleaser.yml`) and a GitHub Actions workflow (`.github/workflows/release.yml`) triggered on `v*` tags. To cut a release: `git tag v0.X.0 && git push --tags`. CI builds and uploads binaries for darwin/linux/windows × amd64/arm64 (windows/arm64 is skipped) and pushes an updated formula to the `TheAndruu/homebrew-tap` repo (`master` branch, `Formula/` directory) so `brew install TheAndruu/tap/ai-clean` picks up the new version. The Homebrew push needs the `HOMEBREW_TAP_GITHUB_TOKEN` secret; binary uploads use the default `GITHUB_TOKEN`. `release.replace_existing_artifacts: true` lets a tag be re-released without manual cleanup.

Shell completions (`completions/ai-clean.{bash,zsh,fish}`) ship inside every release archive and the Homebrew formula installs them into the right Cellar paths. The `curl` and PowerShell install paths in the README don't surface them — only Homebrew users get tab-completion automatically.

Recommended install path on macOS is the Homebrew tap (avoids the Gatekeeper warning on the unsigned binary). The README still documents `curl`/PowerShell one-liners and `go install` as alternatives.

## Where things live

| File | Purpose |
|---|---|
| `main.go` | CLI flags, clipboard I/O, Linux helper detection, `--explain` formatting. `run(args, in, out, err) int` is the testable entry point |
| `main_test.go` | CLI integration tests for the `--stdin` path, flag conflicts, `--explain`, `--version` |
| `internal/clean/clean.go` | `Clean()` orchestrator, `Opts`, `Stats`, fix-point loop, CRLF/CR normalization, blank-line collapse |
| `internal/clean/ansi.go` | Opt-in ANSI/OSC strip |
| `internal/clean/box.go` | Full-box border pre-pass (strips `┌─┐`/`└─┘`/`═══` framing) |
| `internal/clean/leading.go` | Leading whitespace dedent + leading border-char strip (loops to convergence; markdown-table guard) |
| `internal/clean/trailing.go` | Trailing border-char strip + trailing whitespace trim (loops to convergence; markdown-table guard) |
| `internal/clean/rejoin.go` | Wrapped-line rejoin heuristic + table-row guard + unclosed-fence detection |
| `internal/clean/clean_test.go` | Table-driven behavior tests, `TestCleanStats` for the stats fields, `TestCleanIdempotentOverTestdata`, testdata-driven full-pipeline tests |
| `internal/clean/fuzz_test.go` | `FuzzClean` — idempotency property test |
| `internal/clean/bench_test.go` | `BenchmarkClean` — per-fixture throughput baseline |
| `completions/ai-clean.{bash,zsh,fish}` | Static shell-completion scripts, shipped in release archives and installed by the Homebrew formula |
| `testdata/examples/*_sample.txt` | Real captured input for full-pipeline regression cases |
| `testdata/examples/*_expected.txt` | Expected output for the matching sample (regenerable via `go test -update`) |
