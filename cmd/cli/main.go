package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const version = "1.0.0"

type Config struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	Name    string `yaml:"name"`
	Host    string `yaml:"host"`
	BaseDir string `yaml:"base_dir"`
	Default bool   `yaml:"default,omitempty"`
}

type RemoteInfo struct {
	Host    string
	BaseDir string
	App     string
	Env     string
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "server":
		handleServer(os.Args[2:])
	case "apps":
		handleApps(os.Args[2:])
	case "config":
		handleConfig(os.Args[2:])
	case "run":
		handleRun(os.Args[2:])
	case "logs":
		handleLogs(os.Args[2:])
	case "status":
		handleStatus(os.Args[2:])
	case "deploy":
		handleDeploy(os.Args[2:])
	case "rollback":
		handleRollback(os.Args[2:])
	case "restart":
		handleRestart(os.Args[2:])
	case "ssh":
		handleSSH(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("gokku version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'gokku --help' for usage")
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`gokku - Deployment management CLI

Usage:
  gokku <command> [options]

Commands:
  server         Manage servers
  apps           List applications
  config         Manage environment variables
  run            Run arbitrary command
  logs           View application logs
  status         Check services status
  restart        Restart service
  deploy         Deploy applications
  rollback       Rollback to previous release
  ssh            SSH to server
  version        Show version
  help           Show this help

Server Management:
  gokku server add <name> <host>           Add a server
  gokku server list                        List servers
  gokku server remove <name>               Remove a server
  gokku server set-default <name>          Set default server

Configuration Management:
  # Remote execution (from local machine or CI/CD)
  gokku config set KEY=VALUE --remote <git-remote>
  gokku config get KEY --remote <git-remote>
  gokku config list --remote <git-remote>
  gokku config unset KEY --remote <git-remote>

  # Local execution (on server)
  gokku config set KEY=VALUE --app <app> [--env <env>]
  gokku config set KEY=VALUE -a <app> [-e <env>]     (shorthand, env defaults to 'default')
  gokku config get KEY -a <app>                      (uses 'default' env)
  gokku config list -a <app> -e production           (explicit env)
  gokku config unset KEY -a <app>

Run Commands:
  gokku run <command> --remote <git-remote>          Run on remote server
  gokku run <command> -a <app> -e <env>              Run locally

Logs & Status:
  gokku logs <app> <env> [-f] [--remote <git-remote>]
  gokku status [app] [env] [--remote <git-remote>]
  gokku restart <app> <env> [--remote <git-remote>]

Deployment:
  gokku deploy <app> <env> [--remote <git-remote>]
  gokku rollback <app> <env> [--remote <git-remote>]

Examples:
  # Setup server
  gokku server add prod ubuntu@ec2.compute.amazonaws.com

  # Setup git remote (standard git)
  git remote add api-production ubuntu@server:/opt/gokku/repos/api.git
  git remote add vad-staging ubuntu@server:/opt/gokku/repos/vad.git

  # Configuration
  gokku config set PORT=8080 --remote api-production
  gokku config set DATABASE_URL=postgres://... --remote api-production
  gokku config list --remote api-production
  gokku config get PORT --remote vad-staging

  # Run commands
  gokku run "systemctl status api-production" --remote api-production
  gokku run "docker ps" --remote vad-production
  gokku run "bundle exec bin/console" --remote app-production

  # Logs and status
  gokku logs api production -f
  gokku logs --remote api-production -f
  gokku status --remote api-production
  gokku restart --remote vad-staging

  # Deploy
  gokku deploy api production
  gokku deploy --remote api-production

Remote Format:
  --remote <git-remote-name>

  The git remote name (e.g., "api-production", "vad-staging")
  Gokku will run 'git remote get-url <name>' to extract:
  - SSH host (user@ip or user@hostname)
  - App name from path (/opt/gokku/repos/<app>.git)

  Examples of git remotes:
  - api-production → ubuntu@server:/opt/gokku/repos/api.git
  - vad-staging    → ubuntu@server:/opt/gokku/repos/vad.git

  Environment is extracted from remote name suffix:
  - api-production → app: api, env: production
  - vad-staging    → app: vad, env: staging
  - worker-dev     → app: worker, env: dev

Configuration:
  Config file: ~/.gokku/config.yml`)
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gokku", "config.yml")
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Servers: []Server{}}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	configPath := getConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getDefaultServer(config *Config) *Server {
	for _, server := range config.Servers {
		if server.Default {
			return &server
		}
	}
	if len(config.Servers) > 0 {
		return &config.Servers[0]
	}
	return nil
}

// getRemoteInfo extracts info from git remote
// Example: ubuntu@server:/opt/gokku/repos/api.git
// Returns: RemoteInfo{Host: "ubuntu@server", BaseDir: "/opt/gokku", App: "api"}
func getRemoteInfo(remoteName string) (*RemoteInfo, error) {
	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git remote '%s' not found. Add it with: git remote add %s user@host:/opt/gokku/repos/<app>.git", remoteName, remoteName)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse: user@host:/path/to/repos/app.git
	parts := strings.Split(remoteURL, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid remote URL format: %s (expected user@host:/path)", remoteURL)
	}

	host := parts[0]
	path := parts[1]

	// Extract app name from path: /opt/gokku/repos/api.git -> api
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid remote path: %s", path)
	}

	appFile := pathParts[len(pathParts)-1]         // api.git
	appName := strings.TrimSuffix(appFile, ".git") // api

	// Extract base dir: /opt/gokku/repos/api.git -> /opt/gokku
	baseDir := strings.TrimSuffix(path, "/repos/"+appFile)

	// Extract environment from remote name
	// api-production -> production
	// vad-staging -> staging
	// worker-dev -> dev
	env := "production" // default
	nameParts := strings.Split(remoteName, "-")
	if len(nameParts) >= 2 {
		env = nameParts[len(nameParts)-1]
	}

	return &RemoteInfo{
		Host:    host,
		BaseDir: baseDir,
		App:     appName,
		Env:     env,
	}, nil
}

// extractRemoteFlag extracts --remote value from args and returns remaining args
func extractRemoteFlag(args []string) (string, []string) {
	var remote string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--remote" && i+1 < len(args) {
			remote = args[i+1]
			i++ // Skip next arg
		} else {
			remaining = append(remaining, args[i])
		}
	}

	return remote, remaining
}

func handleServer(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku server <add|list|remove|set-default>")
		os.Exit(1)
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "add":
		if len(args) < 3 {
			fmt.Println("Usage: gokku server add <name> <host>")
			os.Exit(1)
		}
		name := args[1]
		host := args[2]

		// Check if server already exists
		for _, s := range config.Servers {
			if s.Name == name {
				fmt.Printf("Server '%s' already exists\n", name)
				os.Exit(1)
			}
		}

		server := Server{
			Name:    name,
			Host:    host,
			BaseDir: "/opt/gokku",
			Default: len(config.Servers) == 0, // First server is default
		}
		config.Servers = append(config.Servers, server)

		if err := saveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Server '%s' added\n", name)
		if server.Default {
			fmt.Println("  Set as default server")
		}

	case "list":
		if len(config.Servers) == 0 {
			fmt.Println("No servers configured")
			fmt.Println("\nAdd a server:")
			fmt.Println("  gokku server add production ubuntu@ec2.compute.amazonaws.com")
			return
		}

		fmt.Println("Configured servers:")
		for _, server := range config.Servers {
			defaultMarker := ""
			if server.Default {
				defaultMarker = " (default)"
			}
			fmt.Printf("  • %s: %s%s\n", server.Name, server.Host, defaultMarker)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: gokku server remove <name>")
			os.Exit(1)
		}
		name := args[1]

		found := false
		newServers := []Server{}
		for _, s := range config.Servers {
			if s.Name != name {
				newServers = append(newServers, s)
			} else {
				found = true
			}
		}

		if !found {
			fmt.Printf("Server '%s' not found\n", name)
			os.Exit(1)
		}

		config.Servers = newServers
		if err := saveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Server '%s' removed\n", name)

	case "set-default":
		if len(args) < 2 {
			fmt.Println("Usage: gokku server set-default <name>")
			os.Exit(1)
		}
		name := args[1]

		found := false
		for i := range config.Servers {
			if config.Servers[i].Name == name {
				config.Servers[i].Default = true
				found = true
			} else {
				config.Servers[i].Default = false
			}
		}

		if !found {
			fmt.Printf("Server '%s' not found\n", name)
			os.Exit(1)
		}

		if err := saveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ '%s' set as default server\n", name)

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleApps(args []string) {
	remote, _ := extractRemoteFlag(args)

	if remote != "" {
		fmt.Printf("Note: --remote flag ignored for 'apps' command\n\n")
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	server := getDefaultServer(config)
	if server == nil {
		fmt.Println("No servers configured")
		fmt.Println("Add a server: gokku server add production ubuntu@ec2.compute.amazonaws.com")
		os.Exit(1)
	}

	fmt.Printf("Listing apps on %s...\n", server.Name)

	cmd := exec.Command("ssh", server.Host, fmt.Sprintf("ls -1 %s/repos 2>/dev/null | sed 's/.git//'", server.BaseDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func envSet(envFile string, pairs []string) {
	envVars := loadEnvFile(envFile)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: invalid format '%s', expected KEY=VALUE\n", pair)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
		fmt.Printf("%s=%s\n", key, value)
	}

	if err := saveEnvFile(envFile, envVars); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}
}

func envGet(envFile string, key string) {
	envVars := loadEnvFile(envFile)

	if value, ok := envVars[key]; ok {
		fmt.Println(value)
	} else {
		fmt.Printf("Error: variable '%s' not found\n", key)
		os.Exit(1)
	}
}

func envList(envFile string) {
	envVars := loadEnvFile(envFile)

	if len(envVars) == 0 {
		fmt.Println("No environment variables set")
		return
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}

	// Sort alphabetically
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, envVars[key])
	}
}

func envUnset(envFile string, keys []string) {
	envVars := loadEnvFile(envFile)

	for _, key := range keys {
		if _, ok := envVars[key]; ok {
			delete(envVars, key)
			fmt.Printf("Unset %s\n", key)
		} else {
			fmt.Printf("Warning: variable '%s' not found\n", key)
		}
	}

	if err := saveEnvFile(envFile, envVars); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}
}

func loadEnvFile(envFile string) map[string]string {
	envVars := make(map[string]string)

	content, err := os.ReadFile(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			return envVars // Return empty map if file doesn't exist
		}
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	return envVars
}

func saveEnvFile(envFile string, envVars map[string]string) error {
	// Sort keys
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var content strings.Builder
	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, envVars[key]))
	}

	return os.WriteFile(envFile, []byte(content.String()), 0600)
}

func handleConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku config <set|get|list|unset> [KEY[=VALUE]] [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <git-remote>     Execute on remote server via SSH")
		fmt.Println("  --app, -a <app>           App name (required for local execution)")
		fmt.Println("  --env, -e <env>           Environment name (optional, defaults to 'default')")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  # Remote execution (from local machine)")
		fmt.Println("  gokku config set PORT=8080 --remote api-production")
		fmt.Println("  gokku config list --remote api-production")
		fmt.Println("")
		fmt.Println("  # Local execution (on server)")
		fmt.Println("  gokku config set PORT=8080 --app api")
		fmt.Println("  gokku config set PORT=8080 -a api                     (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api -e production       (explicit env)")
		fmt.Println("  gokku config list -a api                              (uses 'default' env)")
		os.Exit(1)
	}

	remote, remainingArgs := extractRemoteFlag(args)

	// If --remote is provided, execute via SSH
	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 1 {
			fmt.Println("Usage: gokku config <set|get|list|unset> [args...] --remote <git-remote>")
			os.Exit(1)
		}

		subcommand := remainingArgs[0]

		// Build command to run on server
		var sshCmd string
		switch subcommand {
		case "set":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] --remote <git-remote>")
				os.Exit(1)
			}
			pairs := strings.Join(remainingArgs[1:], " ")
			sshCmd = fmt.Sprintf("gokku config set %s --app %s --env %s", pairs, remoteInfo.App, remoteInfo.Env)
		case "get":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config get KEY --remote <git-remote>")
				os.Exit(1)
			}
			key := remainingArgs[1]
			sshCmd = fmt.Sprintf("gokku config get %s --app %s --env %s", key, remoteInfo.App, remoteInfo.Env)
		case "list":
			sshCmd = fmt.Sprintf("gokku config list --app %s --env %s", remoteInfo.App, remoteInfo.Env)
		case "unset":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config unset KEY [KEY2...] --remote <git-remote>")
				os.Exit(1)
			}
			keys := strings.Join(remainingArgs[1:], " ")
			sshCmd = fmt.Sprintf("gokku config unset %s --app %s --env %s", keys, remoteInfo.App, remoteInfo.Env)
		default:
			fmt.Printf("Unknown subcommand: %s\n", subcommand)
			os.Exit(1)
		}

		fmt.Printf("→ %s/%s (%s)\n", remoteInfo.App, remoteInfo.Env, remoteInfo.Host)

		cmd := exec.Command("ssh", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		return
	}

	// Local execution - parse --app and --env flags
	var appName, envName string
	var finalArgs []string

	for i := 0; i < len(remainingArgs); i++ {
		if (remainingArgs[i] == "--app" || remainingArgs[i] == "-a") && i+1 < len(remainingArgs) {
			appName = remainingArgs[i+1]
			i++
		} else if (remainingArgs[i] == "--env" || remainingArgs[i] == "-e") && i+1 < len(remainingArgs) {
			envName = remainingArgs[i+1]
			i++
		} else {
			finalArgs = append(finalArgs, remainingArgs[i])
		}
	}

	// If no app specified, error
	if appName == "" {
		fmt.Println("Error: --app is required for local execution")
		fmt.Println("")
		fmt.Println("Usage: gokku config <command> [args...] --app <app> [--env <env>]")
		fmt.Println("   or: gokku config <command> [args...] -a <app> [-e <env>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku config set PORT=8080 --app api                    (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api                        (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api -e production          (explicit env)")
		fmt.Println("  gokku config list -a api                                 (uses 'default' env)")
		os.Exit(1)
	}

	// Default environment if not specified
	if envName == "" {
		envName = "default"
	}

	if len(finalArgs) < 1 {
		fmt.Println("Error: command is required (set, get, list, unset)")
		os.Exit(1)
	}

	command := finalArgs[0]

	// Determine env file path
	baseDir := "/opt/gokku"
	if envVar := os.Getenv("GOKKU_BASE_DIR"); envVar != "" {
		baseDir = envVar
	}

	envFile := filepath.Join(baseDir, "apps", appName, envName, ".env")

	// Ensure directory exists
	envDir := filepath.Dir(envFile)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "set":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] --app <app> --env <env>")
			os.Exit(1)
		}
		envSet(envFile, finalArgs[1:])
	case "get":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config get KEY --app <app> --env <env>")
			os.Exit(1)
		}
		envGet(envFile, finalArgs[1])
	case "list":
		envList(envFile)
	case "unset":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] --app <app> --env <env>")
			os.Exit(1)
		}
		envUnset(envFile, finalArgs[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available: set, get, list, unset")
		os.Exit(1)
	}
}

func handleRun(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	if remote == "" {
		fmt.Println("Error: --remote flag is required")
		fmt.Println("Usage: gokku run <command> --remote <git-remote>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku run \"systemctl status api-production\" --remote api-production")
		fmt.Println("  gokku run \"docker ps\" --remote vad-production")
		fmt.Println("  gokku run \"bundle exec bin/console\" --remote app-production")
		os.Exit(1)
	}

	remoteInfo, err := getRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(remainingArgs) < 1 {
		fmt.Println("Error: command is required")
		fmt.Println("Usage: gokku run <command> --remote <git-remote>")
		os.Exit(1)
	}

	// Join all remaining args as the command
	command := strings.Join(remainingArgs, " ")

	fmt.Printf("→ %s/%s (%s)\n", remoteInfo.App, remoteInfo.Env, remoteInfo.Host)
	fmt.Printf("$ %s\n\n", command)

	cmd := exec.Command("ssh", "-t", remoteInfo.Host, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func handleLogs(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var app, env, host string
	var follow bool

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host

		// Check for -f flag
		for _, arg := range remainingArgs {
			if arg == "-f" {
				follow = true
				break
			}
		}
	} else {
		// Legacy: parse from positional args
		if len(remainingArgs) < 2 {
			fmt.Println("Usage: gokku logs <app> <env> [-f]")
			fmt.Println("   or: gokku logs --remote <git-remote> [-f]")
			os.Exit(1)
		}
		app = remainingArgs[0]
		env = remainingArgs[1]
		follow = len(remainingArgs) > 2 && remainingArgs[2] == "-f"

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	followFlag := ""
	if follow {
		followFlag = "-f"
	}

	// Try systemd logs first, fallback to docker logs
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo journalctl -u %s %s -n 100
		elif docker ps -a | grep -q %s; then
			docker logs %s %s
		else
			echo "Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, followFlag, serviceName, followFlag, serviceName, serviceName)

	cmd := exec.Command("ssh", "-t", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func handleStatus(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var app, env, host string

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
	} else {
		// All services
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}

		fmt.Printf("Checking status on %s...\n\n", server.Name)

		sshCmd := fmt.Sprintf(`
			echo "==> Systemd Services"
			for svc in $(ls %s/repos/*.git 2>/dev/null | xargs -n1 basename | sed 's/.git//'); do
				sudo systemctl list-units --all | grep $svc- | awk '{print "  " $1, $3, $4}'
			done
			echo ""
			echo "==> Docker Containers"
			docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null | grep -E "production|staging|develop" || echo "  No containers found"
		`, server.BaseDir)

		cmd := exec.Command("ssh", server.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)

	// Check systemd or docker
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo systemctl status %s
		elif docker ps -a | grep -q %s; then
			docker ps -a | grep %s
		else
			echo "Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, serviceName, serviceName, serviceName)

	cmd := exec.Command("ssh", "-t", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleRestart(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var app, env, host string

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
	} else {
		fmt.Println("Usage: gokku restart <app> <env>")
		fmt.Println("   or: gokku restart --remote <git-remote>")
		os.Exit(1)
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	fmt.Printf("Restarting %s...\n", serviceName)

	// Check systemd or docker and restart accordingly
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo systemctl restart %s && echo "✓ Service restarted"
		elif docker ps -a | grep -q %s; then
			docker restart %s && echo "✓ Container restarted"
		else
			echo "Error: Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, serviceName, serviceName, serviceName)

	cmd := exec.Command("ssh", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleDeploy(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var app, env, remoteName string

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		remoteName = remote
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]
		remoteName = fmt.Sprintf("%s-%s", app, env)
	} else {
		fmt.Println("Usage: gokku deploy <app> <env>")
		fmt.Println("   or: gokku deploy --remote <git-remote>")
		os.Exit(1)
	}

	// Determine branch based on environment
	branch := "main"
	if env == "staging" {
		branch = "staging"
	} else if env == "develop" {
		branch = "develop"
	}

	// Check if remote exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Error: git remote '%s' not found\n", remoteName)
		fmt.Println("\nAdd it with:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remoteName, app)
		os.Exit(1)
	}

	fmt.Printf("Deploying %s to %s environment...\n", app, env)
	fmt.Printf("Remote: %s\n", remoteName)
	fmt.Printf("Branch: %s\n\n", branch)

	// Push
	cmd := exec.Command("git", "push", remoteName, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nDeploy failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Deploy complete!")
}

func handleRollback(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var app, env, host, baseDir string
	var releaseID string

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
		baseDir = remoteInfo.BaseDir

		if len(remainingArgs) > 0 {
			releaseID = remainingArgs[0]
		}
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]
		if len(remainingArgs) > 2 {
			releaseID = remainingArgs[2]
		}

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
		baseDir = server.BaseDir
	} else {
		fmt.Println("Usage: gokku rollback <app> <env> [release-id]")
		fmt.Println("   or: gokku rollback --remote <git-remote> [release-id]")
		os.Exit(1)
	}

	appDir := fmt.Sprintf("%s/apps/%s/%s", baseDir, app, env)
	serviceName := fmt.Sprintf("%s-%s", app, env)

	if releaseID == "" {
		// Get previous release
		listCmd := fmt.Sprintf("cd %s/releases && ls -t | sed -n '2p'", appDir)
		cmd := exec.Command("ssh", host, listCmd)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Failed to get releases: %v\n", err)
			os.Exit(1)
		}
		releaseID = strings.TrimSpace(string(output))
	}

	if releaseID == "" {
		fmt.Println("No previous release found")
		os.Exit(1)
	}

	fmt.Printf("Rolling back %s (%s) to release: %s\n", app, env, releaseID)

	rollbackCmd := fmt.Sprintf(`
		cd %s && \
		if sudo systemctl list-units --all | grep -q %s; then
			sudo systemctl stop %s && \
			ln -sfn %s/releases/%s current && \
			sudo systemctl start %s && \
			echo "✓ Rollback complete"
		elif docker ps -a | grep -q %s; then
			docker stop %s && \
			docker rm %s && \
			docker run -d --name %s --env-file %s/.env -p 8080:8080 %s:release-%s && \
			echo "✓ Rollback complete"
		else
			echo "Error: Service or container not found"
			exit 1
		fi
	`, appDir, serviceName, serviceName, appDir, releaseID, serviceName, serviceName, serviceName, serviceName, serviceName, appDir, app, releaseID)

	cmd := exec.Command("ssh", host, rollbackCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Rollback failed: %v\n", err)
		os.Exit(1)
	}
}

func handleSSH(args []string) {
	remote, remainingArgs := extractRemoteFlag(args)

	var host string

	if remote != "" {
		remoteInfo, err := getRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		host = remoteInfo.Host
		fmt.Printf("Connecting to %s (%s/%s)...\n", host, remoteInfo.App, remoteInfo.Env)
	} else {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := getDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
		fmt.Printf("Connecting to %s...\n", server.Name)
	}

	cmd := exec.Command("ssh", append([]string{"-t", host}, remainingArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}
