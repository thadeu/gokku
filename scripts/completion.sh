#!/bin/bash
# Gokku bash/zsh completion script
# Supports both server mode (apps from /opt/gokku/apps) and client mode (git remotes)
# Error handling ensures completion never breaks the shell

_gokku() {
    # Disable error exit for completion function
    set +e

    local cur prev words cword
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
        cword=${CURRENT}
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
        cword=${COMP_CWORD}
    fi

    # Main commands (including commands with colons)
    cmds="apps config deploy logs ps ps:scale ps:list ps:restart ps:stop ps:report restart rollback run server services services:list services:create services:link services:unlink services:destroy services:info services:logs ssh status tool plugins plugins:list plugins:install plugins:uninstall config:set config:get config:list config:unset version help au update auto-update"

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

    # Handle ps command - complete with ps: commands
    if [[ "$prev" == "ps" ]] || [[ "$cur" == ps:* ]]; then
        local ps_subcommands="ps:scale ps:list ps:restart ps:stop ps:report"
        COMPREPLY=($(compgen -W "$ps_subcommands" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle services command
    if [[ "$prev" == "services" ]] || [[ "$cur" == services:* ]]; then
        local services_subcommands="services:list services:create services:link services:unlink services:destroy services:info services:logs"
        COMPREPLY=($(compgen -W "$services_subcommands" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle plugins command
    if [[ "$prev" == "plugins" ]] || [[ "$cur" == plugins:* ]]; then
        local plugins_subcommands="plugins:list plugins:install plugins:uninstall"
        COMPREPLY=($(compgen -W "$plugins_subcommands" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle config command
    if [[ "$prev" == "config" ]] || [[ "$cur" == config:* ]]; then
        local config_subcommands="config:set config:get config:list config:unset"
        COMPREPLY=($(compgen -W "$config_subcommands" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'apps'
    if [[ "$prev" == "apps" ]]; then
        COMPREPLY=($(compgen -W "list ls create destroy rm" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Subcommands for 'server'
    if [[ "$prev" == "server" ]]; then
        COMPREPLY=($(compgen -W "add list remove set-default" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle ps:scale, ps:list, etc - they need -a flag
    if [[ "$prev" =~ ^ps:(scale|list|restart|stop|report)$ ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle services: commands that need -a flag
    if [[ "$prev" =~ ^services:(link|unlink)$ ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Handle config: commands that need -a flag
    if [[ "$prev" =~ ^config:(set|get|list|unset)$ ]]; then
        COMPREPLY=($(compgen -W "-a --app" -- "$cur" 2>/dev/null || true))
        return 0
    fi

    # Default: complete with main commands (first argument only)
    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$cmds" -- "$cur" 2>/dev/null || true))
        return 0
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

