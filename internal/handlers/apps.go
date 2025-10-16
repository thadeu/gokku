package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleApps manages applications on the server
func handleApps(args []string) {
	if len(args) == 0 {
		handleAppsList()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		handleAppsList()
	case "create":
		handleAppsCreate(args[1:])
	case "destroy", "rm":
		handleAppsDestroy(args[1:])
	default:
		fmt.Println("Usage: gokku apps <command> [options]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  list, ls              List all applications")
		fmt.Println("  create <app> [env]    Create application and setup deployment")
		fmt.Println("  destroy, rm <app>     Destroy application")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <remote>     Use specific git remote")
		os.Exit(1)
	}
}

// handleAppsList lists applications on the server
func handleAppsList() {
	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	server := internal.GetDefaultServer(config)
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

// handleAppsCreate creates an application and sets up deployment
func handleAppsCreate(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps create <app> [environment] [--remote <remote>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku apps create myapp                    # uses default environment")
		fmt.Println("  gokku apps create myapp production         # explicit environment")
		fmt.Println("  gokku apps create myapp --remote myremote  # uses git remote")
		os.Exit(1)
	}

	appName := remainingArgs[0]
	envName := "production" // default

	if len(remainingArgs) >= 2 {
		envName = remainingArgs[1]
	}

	if remote == "" {
		// Try to find a remote that matches the app name pattern
		remote = fmt.Sprintf("%s-%s", appName, envName)
		fmt.Printf("Using remote: %s\n", remote)
	}

	// Parse remote to get connection info
	remoteInfo, err := internal.GetRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("")
		fmt.Println("Make sure the git remote exists:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remote, appName)
		os.Exit(1)
	}

	fmt.Printf("Creating app %s (%s) on %s...\n", appName, envName, remoteInfo.Host)

	// Setup repository automatically
	if err := autoSetupRepository(remoteInfo); err != nil {
		fmt.Printf("Failed to setup repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ App created successfully!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Make sure your gokku.yml is committed")
	fmt.Println("  2. Deploy with: git push", remote, "main")
}

// handleAppsDestroy destroys an application
func handleAppsDestroy(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps destroy <app> [--remote <remote>]")
		os.Exit(1)
	}

	appName := remainingArgs[0]

	if remote == "" {
		fmt.Println("Error: --remote is required for destroy command")
		fmt.Println("This prevents accidental deletion of the wrong app")
		os.Exit(1)
	}

	remoteInfo, err := internal.GetRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Destroying app %s on %s...\n", appName, remoteInfo.Host)
	fmt.Printf("This will permanently delete the app and all its data.\n")
	fmt.Printf("Continue? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Aborted.")
		return
	}

	// Remove app directory and repository
	destroyCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		echo "Removing app directory..."
		sudo rm -rf /opt/gokku/apps/%s
		echo "Removing repository..."
		sudo rm -rf /opt/gokku/repos/%s.git
		echo "App destroyed successfully"
	`, appName, appName))

	destroyCmd.Stdout = os.Stdout
	destroyCmd.Stderr = os.Stderr

	if err := destroyCmd.Run(); err != nil {
		fmt.Printf("Failed to destroy app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ App destroyed successfully!")
}

// autoSetupRepository attempts to create the repository on the server if it doesn't exist
func autoSetupRepository(remoteInfo *internal.RemoteInfo) error {
	fmt.Printf("Checking repository status on %s...\n", remoteInfo.Host)

	// Test SSH connection and check if repo exists
	testCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("test -d /opt/gokku/repos/%s.git", remoteInfo.App))
	if err := testCmd.Run(); err == nil {
		fmt.Println("✓ Repository exists")
		return nil
	}

	fmt.Printf("Repository doesn't exist, creating /opt/gokku/repos/%s.git...\n", remoteInfo.App)

	// Get the user from SSH connection (more reliable than os.Getenv)
	userCmd := exec.Command("ssh", remoteInfo.Host, "whoami")
	userOutput, err := userCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get user from server: %v", err)
	}
	serverUser := strings.TrimSpace(string(userOutput))

	// Create repository on server
	setupCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		echo "Creating repository directory..."
		sudo mkdir -p /opt/gokku/repos
		sudo mkdir -p /opt/gokku/repos/%s.git
		echo "Initializing git repository..."
		sudo git init --bare /opt/gokku/repos/%s.git
		echo "Setting permissions..."
		sudo chown -R %s /opt/gokku/repos/%s.git
		echo "Repository created successfully"
	`, remoteInfo.App, remoteInfo.App, serverUser, remoteInfo.App))

	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr

	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	// Copy the smart hook from local Gokku installation
	// First try to find the hook in the local Gokku installation
	localHookPaths := []string{
		"./hooks/post-receive-systemd.template",                      // If running from source
		"/usr/local/share/gokku/hooks/post-receive-systemd.template", // Standard install location
		"/opt/gokku/hooks/post-receive-systemd.template",             // Alternative location
	}

	var hookPath string
	for _, path := range localHookPaths {
		if _, err := os.Stat(path); err == nil {
			hookPath = path
			break
		}
	}

	if hookPath == "" {
		// Create a smart hook that calls gokku deploy on server
		fmt.Println("Creating smart hook that delegates to server Gokku...")
		configCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
			cat > /opt/gokku/repos/%s.git/hooks/post-receive << 'EOF'
#!/bin/bash
set -e

APP_NAME="%s"
ENVIRONMENT="production"
BASE_DIR="/opt/gokku"
APP_DIR="$BASE_DIR/apps/$APP_NAME/$ENVIRONMENT"
REPO_DIR="$BASE_DIR/repos/$APP_NAME.git"
RELEASE_DIR="$APP_DIR/releases/$(date +%%Y%%m%%d-%%H%%M%%S)"
SERVICE_NAME="$APP_NAME-$ENVIRONMENT"
BINARY_NAME="$APP_NAME"

# Default values (will be overridden by gokku.yml if available)
BUILD_TYPE="systemd"
LANG="go"
BUILD_PATH="./cmd/$APP_NAME"
BUILD_WORKDIR="."
KEEP_RELEASES="5"

# Source mise helpers if available
SCRIPT_DIR="/opt/gokku"
if [ -f "$SCRIPT_DIR/mise-helpers.sh" ]; then
    source "$SCRIPT_DIR/mise-helpers.sh"
fi

# Router for environment-specific hooks
while read oldrev newrev refname; do
    branch=$(git rev-parse --symbolic --abbrev-ref $refname)
    echo "==> Received push to branch: $branch"

    if [[ "$branch" == "main" ]]; then
        echo "==> Deploying to production environment..."
        break
    fi
done

echo "-----> Deploying $APP_NAME to $ENVIRONMENT..."

# Create release directory
mkdir -p "$RELEASE_DIR"

# Extract code - try main first, then HEAD
echo "-----> Extracting code..."
if GIT_WORK_TREE="$RELEASE_DIR" git checkout -f main 2>/dev/null; then
    echo "-----> Checked out main branch"
elif GIT_WORK_TREE="$RELEASE_DIR" git checkout -f HEAD 2>/dev/null; then
    echo "-----> Checked out HEAD"
else
    echo "-----> No commits available for checkout, using defaults..."
fi

# Update gokku.yml from repository if it exists
if [ -f "$RELEASE_DIR/gokku.yml" ]; then
    echo "-----> Updating gokku.yml from repository..."
    cp -f "$RELEASE_DIR/gokku.yml" "$SCRIPT_DIR/gokku.yml" 2>/dev/null || echo "No gokku.yml found"

    # Try to read app-specific configuration from gokku.yml
    if command -v yq >/dev/null 2>&1; then
        echo "-----> Reading app configuration from gokku.yml..."

        # Read build configuration for this app
        if yq ".apps[] | select(.name == \"$APP_NAME\")" "$RELEASE_DIR/gokku.yml" >/dev/null 2>&1; then
            BUILD_TYPE=$(yq ".apps[] | select(.name == \"$APP_NAME\") | .build.type // \"$BUILD_TYPE\"" "$RELEASE_DIR/gokku.yml" 2>/dev/null || echo "$BUILD_TYPE")
            LANG=$(yq ".apps[] | select(.name == \"$APP_NAME\") | .lang // \"$LANG\"" "$RELEASE_DIR/gokku.yml" 2>/dev/null || echo "$LANG")
            BUILD_PATH=$(yq ".apps[] | select(.name == \"$APP_NAME\") | .build.path // \"$BUILD_PATH\"" "$RELEASE_DIR/gokku.yml" 2>/dev/null | tr -d '"' || echo "$BUILD_PATH")
            BUILD_WORKDIR=$(yq ".apps[] | select(.name == \"$APP_NAME\") | .build.work_dir // \"$BUILD_WORKDIR\"" "$RELEASE_DIR/gokku.yml" 2>/dev/null | tr -d '"' || echo "$BUILD_WORKDIR")
            KEEP_RELEASES=$(yq ".apps[] | select(.name == \"$APP_NAME\") | .deployment.keep_releases // \"$KEEP_RELEASES\"" "$RELEASE_DIR/gokku.yml" 2>/dev/null || echo "$KEEP_RELEASES")

            echo "==> Config loaded: TYPE=$BUILD_TYPE, LANG=$LANG, WORKDIR=$BUILD_WORKDIR"
        else
            echo "==> App '$APP_NAME' not found in gokku.yml, using defaults"
        fi
    fi

    source "$SCRIPT_DIR/config-loader.sh" 2>/dev/null || echo "Config loader not found"
else
    echo "-----> No gokku.yml found in repository, using defaults"
fi

# Auto-setup if needed
if [ ! -d "$APP_DIR/releases" ] || [ -z "$(ls -A "$APP_DIR/releases" 2>/dev/null)" ]; then
    echo "-----> First deploy detected, delegating to gokku deploy..."
    mkdir -p "$APP_DIR"/{releases,shared}

    # Create .env file
    if [ ! -f "$APP_DIR/shared/.env" ]; then
        cat > "$APP_DIR/shared/.env" << ENV_EOF
# Environment: $ENVIRONMENT
# App: $APP_NAME
# Generated: $(date)

# Add your environment variables here
PORT=8080
ENV_EOF
        echo "==> Created .env file"
    fi

    # Create systemd service
    sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << SERVICE_EOF
[Unit]
Description=$APP_NAME ($ENVIRONMENT)
After=network.target

[Service]
Type=simple
User=thadeu
WorkingDirectory=$APP_DIR/current
ExecStart=$APP_DIR/current/$BINARY_NAME
Restart=always
RestartSec=5

EnvironmentFile=$APP_DIR/shared/.env

[Install]
WantedBy=multi-user.target
SERVICE_EOF

    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    echo "==> Created systemd service"
else
    echo "-----> App already configured, proceeding with normal deploy"
fi

# Setup mise if .tool-versions exists
if [ -f "$RELEASE_DIR/$BUILD_WORKDIR/.tool-versions" ]; then
    echo "-----> Detected .tool-versions, setting up mise..."

    # Install mise if not present
    if ! ~/.local/bin/mise --version >/dev/null 2>&1; then
        echo "-----> Installing mise..."
        curl https://mise.run | sh
    fi

    # Install tools
    cd "$RELEASE_DIR/$BUILD_WORKDIR"
    ~/.local/bin/mise install || exit 1

    # Activate mise
    export PATH="$HOME/.local/bin:$PATH"
    eval "$(~/.local/bin/mise activate bash)"
fi

# Build only if we have source code
if [ -d "$RELEASE_DIR/$BUILD_WORKDIR" ] && [ -f "$RELEASE_DIR/$BUILD_WORKDIR/go.mod" ]; then
    echo "-----> Building $APP_NAME..."
    cd "$RELEASE_DIR/$BUILD_WORKDIR"

    # Go build
    export GOOS=linux
    export GOARCH=arm64
    export CGO_ENABLED=0
    go build -o "$RELEASE_DIR/$BINARY_NAME" $BUILD_PATH

    # Deploy
    echo "-----> Deploying..."
    ln -sf "$RELEASE_DIR" "$APP_DIR/current"
    ln -sf "$APP_DIR/shared/.env" "$RELEASE_DIR/.env"

    # Restart service
    sudo systemctl restart "$SERVICE_NAME"

    echo "-----> Deploy complete!"
else
    echo "-----> No source code found, skipping build and deploy"
    echo "-----> This appears to be the initial repository setup"
fi
EOF
			chmod +x /opt/gokku/repos/%s.git/hooks/post-receive
			echo "Smart hook created successfully"
		`, remoteInfo.App, remoteInfo.App, remoteInfo.App))

		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr

		if err := configCmd.Run(); err != nil {
			return fmt.Errorf("failed to create smart hook: %v", err)
		}
	} else {
		// Copy the actual hook template
		copyCmd := exec.Command("scp", hookPath, fmt.Sprintf("%s:/opt/gokku/repos/%s.git/hooks/post-receive", remoteInfo.Host, remoteInfo.App))
		copyCmd.Stdout = os.Stdout
		copyCmd.Stderr = os.Stderr

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy hook template: %v", err)
		}

		// Make it executable
		execCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("chmod +x /opt/gokku/repos/%s.git/hooks/post-receive", remoteInfo.App))
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("failed to make hook executable: %v", err)
		}
	}

	fmt.Println("✓ Repository auto-setup complete")
	return nil
}
