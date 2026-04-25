# AI Clean

Clean up output you copied from AI CLI tools (Claude Code, GitHub Copilot CLI, and similar) so it pastes into chat, GitHub issues, docs, or another terminal without the usual mess: trailing whitespace padding, terminal-wrapped lines, and left/right border characters like `│` or `|`.

Workflow: **select → copy → `ai-clean` → paste**.

## What it fixes

Before (copied straight from a bordered CLI):

```
  │ Here is a small example showing the structure that copilot's CLI │
  │ produces. Each line has a left and right border character.       │
  │                                                                  │
  │     def greet(name):                                             │
  │         return f"hello, {name}"                                  │
```

After `ai-clean`:

```
Here is a small example showing the structure that copilot's CLI produces. Each line has a left and right border character.

    def greet(name):
        return f"hello, {name}"
```

Code-block indentation is preserved; prose is rejoined; borders and trailing whitespace are gone.

## Install

### macOS / Linux / Windows — prebuilt binary

Download the latest release from the [Releases page](https://github.com/TheAndruu/ai-clean/releases) and put `ai-clean` somewhere on your `$PATH`.

### From source

```sh
go install github.com/TheAndruu/ai-clean@latest
```

### Linux requirement

`ai-clean` shells out to a system clipboard helper on Linux. Install one:

```sh
sudo apt install xclip          # Debian / Ubuntu
sudo pacman -S xclip            # Arch
sudo dnf install xclip          # Fedora
# Wayland: sudo apt install wl-clipboard
```

macOS (`pbcopy`/`pbpaste`) and Windows (`clip.exe`) ship with the OS — no install needed.

## Usage

```
ai-clean              # read clipboard, clean, write back
ai-clean --dry-run    # print cleaned text instead of writing back to clipboard
ai-clean --stdin      # read stdin, write stdout (composable in pipelines)
ai-clean --no-rejoin  # skip the wrapped-line rejoin (safer when pasting pure code)
ai-clean --strip-ansi # also strip ANSI / OSC escape sequences (off by default)
ai-clean --version
```

Examples:

```sh
# Most common: copy from terminal, run, paste into your editor.
ai-clean

# Pipe a captured log file through it.
cat session.log | ai-clean --stdin > clean.txt

# Verify what would happen without modifying the clipboard.
ai-clean --dry-run
```

## How it works

The cleanup runs in a fixed order, designed to be safe across plain prose, plain code, and mixed content:

1. **Optional ANSI / OSC strip** (only with `--strip-ansi`). Off by default because terminals usually strip these on copy already.
2. **Strip leading chrome.** Computes the minimum leading-whitespace count across non-empty lines and dedents — preserving relative indentation of code blocks. Then detects a uniform border character (`│`, `|`, `>`, etc.) appearing on ≥80% of non-empty lines and strips it. Loops until stable, so nested borders (`│ │ text`) peel cleanly.
3. **Strip trailing chrome.** Mirror of step 2 for the right side: detects a uniform trailing border character (looking past trailing whitespace), strips it, then trims trailing whitespace per line. Looped for nested borders.
4. **Rejoin wrapped lines.** Conservative heuristic that merges adjacent prose lines only when all of the following hold: the previous line doesn't end in sentence-terminating punctuation, the next doesn't start with a capital / list marker / heading marker, neither side has leading whitespace, and the document's longest line is at least 40 chars (a proxy for terminal-wrapped output). Skipped entirely inside fenced code blocks. Use `--no-rejoin` to disable.
5. **Cosmetic blank-line collapse.** Runs of 3+ consecutive blank lines collapse to 2.

## License

MIT
