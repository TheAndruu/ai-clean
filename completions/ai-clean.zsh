#compdef ai-clean
_ai_clean() {
    _arguments \
        '--stdin[read text from stdin and write cleaned text to stdout]' \
        '--dry-run[print cleaned text to stdout instead of writing clipboard]' \
        '--no-rejoin[disable wrapped-line rejoin heuristic]' \
        '--strip-ansi[also strip ANSI / OSC escape sequences]' \
        '--explain[print stage-by-stage summary to stderr]' \
        '--version[print version and exit]' \
        '--help[show help]'
}
_ai_clean "$@"
