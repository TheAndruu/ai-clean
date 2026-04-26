_ai_clean_completions() {
    local cur
    cur="${COMP_WORDS[COMP_CWORD]}"
    local flags="--stdin --dry-run --no-rejoin --strip-ansi --explain --version --help"
    COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )
    return 0
}
complete -F _ai_clean_completions ai-clean
