#!/bin/bash
# Gokku bash/zsh completion script
# Supports both server mode (apps from /opt/gokku/apps) and client mode (git remotes)
# Error handling ensures completion never breaks the shell

_gokku() {
    # Disable error exit for completion function
    set +e

    local cur prev
    local cmds
    local is_server_mode=false

    # Safely detect server mode
    if [ -f ~/.gokkurc ] 2>/dev/null; then
        if grep -q "mode=server" ~/.gokkurc 2>/dev/null; then
            is_server_mode=true
        fi
    fi

    # Get current word and previous word (bash/zsh compatible)
    if [ -n "$ZSH_VERSION" ]; then
        cur="${words[CURRENT]}"
        prev="${words[CURRENT-1]}"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    fi

    # Main commands
    cmds="apps config deploy logs ps restart rollback run server services ssh status tool plugins version help au update auto-update"

    # Check if we're completing a flag value
    if [[ "$prev" == "-a" || "$prev" == "--app" ]]; then
        if [ "$is_server_mode" = true ]; then
            # Server mode: list apps from /opt/gokku/apps
            if [ -d /opt/gokku/apps ] 2>/dev/null; then
                local apps=$(ls -1 /opt/gokku/apps 2>/dev/null | grep -v "^$" || true)
                if [ -n "$apps" ]; then
                    COMPREPLY=($(compgen -W "$apps" -- "$cur" 2>/dev/null || true))
                fi
            fi
        else
            # Client mode: list git remotes
            if command -v git >/dev/null 2>&1; then
                local remotes=$(git remote 2>/dev/null || true)
                if [ -n "$remotes" ]; then
                    COMPREPLY=($(compgen -W "$remotes" -- "$cur" 2>/dev/null || true))
                fi
            fi
        fi
        return 0
    fi

    # Check if we're completing after a command that needs -a flag
    local needs_app_flag=("config" "run" "logs" "status" "restart" "rollback")
    if [[ " ${needs_app_flag[@]} " =~ " ${prev} " ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle ps commands
    if [[ "$prev" == "ps" ]]; then
        COMPREPLY=($(compgen -W "scale list restart stop report" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle ps: commands
    if [[ "$prev" =~ ^ps: ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'apps'
    if [[ "$prev" == "apps" ]]; then
        COMPREPLY=($(compgen -W "list ls create destroy rm" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'config'
    if [[ "$prev" == "config" ]]; then
        COMPREPLY=($(compgen -W "set get list unset" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle config: commands
    if [[ "$prev" =~ ^config: ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'server'
    if [[ "$prev" == "server" ]]; then
        COMPREPLY=($(compgen -W "add list remove set-default" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'services'
    if [[ "$prev" == "services" ]]; then
        COMPREPLY=($(compgen -W "list create link unlink destroy info logs" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle services: commands
    if [[ "$prev" =~ ^services: ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'plugins'
    if [[ "$prev" == "plugins" ]]; then
        COMPREPLY=($(compgen -W "list install uninstall" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle plugins: commands
    if [[ "$prev" =~ ^plugins: ]]; then
        # No completion for plugin commands
        return 0
    fi

    # Default: complete with main commands
    if [ -n "$ZSH_VERSION" ]; then
        if [ ${#words[@]} -eq 2 ]; then
            COMPREPLY=($(compgen -W "$cmds" -- "$cur" 2>/dev/null || true))
        fi
    else
        if [[ $COMP_CWORD -eq 1 ]]; then
            COMPREPLY=($(compgen -W "$cmds" -- "$cur" 2>/dev/null || true))
        fi
    fi

    # Ensure COMPREPLY is always set (bash only)
    if [ -z "$ZSH_VERSION" ] && [ -z "${COMPREPLY[*]}" ]; then
        COMPREPLY=()
    fi

    return 0
}

# Register completion for bash
if [ -n "$BASH_VERSION" ]; then
    complete -F _gokku gokku 2>/dev/null || true
fi

# Register completion for zsh using bashcompinit (compatible approach)
if [ -n "$ZSH_VERSION" ]; then
    autoload -U +X compinit 2>/dev/null && compinit 2>/dev/null || true
    autoload -U +X bashcompinit 2>/dev/null && bashcompinit 2>/dev/null || true
    complete -F _gokku gokku 2>/dev/null || true
fi

