# Gokku Fish shell completion
# Supports both server mode (apps from /opt/gokku/apps) and client mode (git remotes)

# Complete main command
complete -c gokku -f

# Main commands
complete -c gokku -n '__fish_use_subcommand' -a 'apps' -d 'List applications'
complete -c gokku -n '__fish_use_subcommand' -a 'config' -d 'Manage environment variables'
complete -c gokku -n '__fish_use_subcommand' -a 'deploy' -d 'Deploy applications'
complete -c gokku -n '__fish_use_subcommand' -a 'logs' -d 'View application logs'
complete -c gokku -n '__fish_use_subcommand' -a 'ps' -d 'Process management'
complete -c gokku -n '__fish_use_subcommand' -a 'restart' -d 'Restart services'
complete -c gokku -n '__fish_use_subcommand' -a 'rollback' -d 'Rollback to previous release'
complete -c gokku -n '__fish_use_subcommand' -a 'run' -d 'Run arbitrary commands'
complete -c gokku -n '__fish_use_subcommand' -a 'server' -d 'Manage server connections'
complete -c gokku -n '__fish_use_subcommand' -a 'services' -d 'Manage services'
complete -c gokku -n '__fish_use_subcommand' -a 'ssh' -d 'SSH to server'
complete -c gokku -n '__fish_use_subcommand' -a 'status' -d 'Check services status'
complete -c gokku -n '__fish_use_subcommand' -a 'tool' -d 'Utility commands'
complete -c gokku -n '__fish_use_subcommand' -a 'plugins' -d 'Manage plugins'
complete -c gokku -n '__fish_use_subcommand' -a 'version' -d 'Show version'
complete -c gokku -n '__fish_use_subcommand' -a 'help' -d 'Show help'
complete -c gokku -n '__fish_use_subcommand' -a 'au' -d 'Auto-update'
complete -c gokku -n '__fish_use_subcommand' -a 'update' -d 'Auto-update'
complete -c gokku -n '__fish_use_subcommand' -a 'auto-update' -d 'Auto-update'

# Commands with colons
complete -c gokku -n '__fish_use_subcommand' -a 'ps:scale' -d 'Scale app processes'
complete -c gokku -n '__fish_use_subcommand' -a 'ps:list' -d 'List running processes'
complete -c gokku -n '__fish_use_subcommand' -a 'ps:restart' -d 'Restart all processes'
complete -c gokku -n '__fish_use_subcommand' -a 'ps:stop' -d 'Stop processes'
complete -c gokku -n '__fish_use_subcommand' -a 'ps:report' -d 'List running processes'

complete -c gokku -n '__fish_use_subcommand' -a 'services:list' -d 'List all services'
complete -c gokku -n '__fish_use_subcommand' -a 'services:create' -d 'Create service from plugin'
complete -c gokku -n '__fish_use_subcommand' -a 'services:link' -d 'Link service to app'
complete -c gokku -n '__fish_use_subcommand' -a 'services:unlink' -d 'Unlink service from app'
complete -c gokku -n '__fish_use_subcommand' -a 'services:destroy' -d 'Destroy service'
complete -c gokku -n '__fish_use_subcommand' -a 'services:info' -d 'Show service information'
complete -c gokku -n '__fish_use_subcommand' -a 'services:logs' -d 'Show service logs'

complete -c gokku -n '__fish_use_subcommand' -a 'plugins:list' -d 'List all installed plugins'
complete -c gokku -n '__fish_use_subcommand' -a 'plugins:install' -d 'Install plugin'
complete -c gokku -n '__fish_use_subcommand' -a 'plugins:uninstall' -d 'Uninstall plugin'

complete -c gokku -n '__fish_use_subcommand' -a 'config:set' -d 'Set environment variable'
complete -c gokku -n '__fish_use_subcommand' -a 'config:get' -d 'Get environment variable'
complete -c gokku -n '__fish_use_subcommand' -a 'config:list' -d 'List environment variables'
complete -c gokku -n '__fish_use_subcommand' -a 'config:unset' -d 'Unset environment variable'

# Apps subcommands
complete -c gokku -n '__fish_seen_subcommand_from apps' -a 'list' -d 'List all applications'
complete -c gokku -n '__fish_seen_subcommand_from apps' -a 'ls' -d 'List all applications'
complete -c gokku -n '__fish_seen_subcommand_from apps' -a 'create' -d 'Create application'
complete -c gokku -n '__fish_seen_subcommand_from apps' -a 'destroy' -d 'Destroy application'
complete -c gokku -n '__fish_seen_subcommand_from apps' -a 'rm' -d 'Destroy application'

# Server subcommands
complete -c gokku -n '__fish_seen_subcommand_from server' -a 'add' -d 'Add a server'
complete -c gokku -n '__fish_seen_subcommand_from server' -a 'list' -d 'List servers'
complete -c gokku -n '__fish_seen_subcommand_from server' -a 'remove' -d 'Remove a server'
complete -c gokku -n '__fish_seen_subcommand_from server' -a 'set-default' -d 'Set default server'

# Flags
complete -c gokku -s a -l app -d 'Specify app name or git remote'

# Function to get apps or remotes based on mode
function __gokku_get_apps_or_remotes
    if test -f ~/.gokkurc
        if grep -q "mode=server" ~/.gokkurc 2>/dev/null
            # Server mode: list apps from /opt/gokku/apps
            if test -d /opt/gokku/apps
                ls -1 /opt/gokku/apps 2>/dev/null | grep -v "^$"
            end
        else
            # Client mode: list git remotes
            if command -v git >/dev/null 2>&1
                git remote 2>/dev/null
            end
        end
    else
        # Default to client mode, list git remotes
        if command -v git >/dev/null 2>&1
            git remote 2>/dev/null
        end
    end
end

# Complete -a/--app flag with apps or remotes
complete -c gokku -n '__fish_seen_argument -s a; or __fish_seen_argument -l app' -a '(__gokku_get_apps_or_remotes)'

# Complete -a/--app after commands that need it
complete -c gokku -n '__fish_seen_subcommand_from config run logs status restart rollback' -s a -l app -a '(__gokku_get_apps_or_remotes)'
complete -c gokku -n '__fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -s a -l app -a '(__gokku_get_apps_or_remotes)'
complete -c gokku -n '__fish_seen_subcommand_from services:link services:unlink' -s a -l app -a '(__gokku_get_apps_or_remotes)'
complete -c gokku -n '__fish_seen_subcommand_from config:set config:get config:list config:unset' -s a -l app -a '(__gokku_get_apps_or_remotes)'

# Complete ps subcommands when typing ps
complete -c gokku -n '__fish_seen_subcommand_from ps; and not __fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -a 'ps:scale' -d 'Scale app processes'
complete -c gokku -n '__fish_seen_subcommand_from ps; and not __fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -a 'ps:list' -d 'List running processes'
complete -c gokku -n '__fish_seen_subcommand_from ps; and not __fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -a 'ps:restart' -d 'Restart all processes'
complete -c gokku -n '__fish_seen_subcommand_from ps; and not __fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -a 'ps:stop' -d 'Stop processes'
complete -c gokku -n '__fish_seen_subcommand_from ps; and not __fish_seen_subcommand_from ps:scale ps:list ps:restart ps:stop ps:report' -a 'ps:report' -d 'List running processes'

# Complete services subcommands when typing services
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:list' -d 'List all services'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:create' -d 'Create service from plugin'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:link' -d 'Link service to app'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:unlink' -d 'Unlink service from app'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:destroy' -d 'Destroy service'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:info' -d 'Show service information'
complete -c gokku -n '__fish_seen_subcommand_from services; and not __fish_seen_subcommand_from services:list services:create services:link services:unlink services:destroy services:info services:logs' -a 'services:logs' -d 'Show service logs'

# Complete plugins subcommands when typing plugins
complete -c gokku -n '__fish_seen_subcommand_from plugins; and not __fish_seen_subcommand_from plugins:list plugins:install plugins:uninstall' -a 'plugins:list' -d 'List all installed plugins'
complete -c gokku -n '__fish_seen_subcommand_from plugins; and not __fish_seen_subcommand_from plugins:list plugins:install plugins:uninstall' -a 'plugins:install' -d 'Install plugin'
complete -c gokku -n '__fish_seen_subcommand_from plugins; and not __fish_seen_subcommand_from plugins:list plugins:install plugins:uninstall' -a 'plugins:uninstall' -d 'Uninstall plugin'

# Complete config subcommands when typing config
complete -c gokku -n '__fish_seen_subcommand_from config; and not __fish_seen_subcommand_from config:set config:get config:list config:unset' -a 'config:set' -d 'Set environment variable'
complete -c gokku -n '__fish_seen_subcommand_from config; and not __fish_seen_subcommand_from config:set config:get config:list config:unset' -a 'config:get' -d 'Get environment variable'
complete -c gokku -n '__fish_seen_subcommand_from config; and not __fish_seen_subcommand_from config:set config:get config:list config:unset' -a 'config:list' -d 'List environment variables'
complete -c gokku -n '__fish_seen_subcommand_from config; and not __fish_seen_subcommand_from config:set config:get config:list config:unset' -a 'config:unset' -d 'Unset environment variable'

