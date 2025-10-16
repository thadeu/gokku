#!/bin/bash
# Generic server setup script for git-push deployment
# Reads configuration from gokku.yml
# Usage: ./deploy-server-setup.sh <app-name> <environment>

set -e

if [ $# -lt 2 ]; then
    echo "Usage: $0 <app-name> <environment>"
    echo ""
    echo "Configuration is read from gokku.yml"
    echo ""
    echo "Examples:"
    echo "  $0 my-app production"
    echo "  $0 api-service staging"
    exit 1
fi

APP_NAME="$1"
ENVIRONMENT="$2"

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${GOKKU_CONFIG:-$SCRIPT_DIR/gokku.yml}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file $CONFIG_FILE not found"
    exit 1
fi

# Source config loader
source "$SCRIPT_DIR/config-loader.sh"

# Validate app exists in config
if ! get_apps | grep -q "^${APP_NAME}$"; then
    echo "Error: App '$APP_NAME' not found in $CONFIG_FILE"
    echo ""
    echo "Available apps:"
    get_apps | sed 's/^/  - /'
    exit 1
fi

# Validate environment exists for this app
if ! get_app_environments "$APP_NAME" | grep -q "^${ENVIRONMENT}$"; then
    echo "Error: Environment '$ENVIRONMENT' not found for app '$APP_NAME' in $CONFIG_FILE"
    echo ""
    echo "Available environments for $APP_NAME:"
    get_app_environments "$APP_NAME" | sed 's/^/  - /'
    exit 1
fi

# Get configuration values
BUILD_TYPE=$(get_app_build_type "$APP_NAME")
LANG=$(get_app_lang "$APP_NAME")
BINARY_NAME=$(get_app_binary_name "$APP_NAME")
BUILD_PATH=$(get_app_build_path "$APP_NAME")
BUILD_WORKDIR=$(get_app_work_dir "$APP_NAME")
ENV_BRANCH=$(get_app_env_branch "$APP_NAME" "$ENVIRONMENT")
DOCKERFILE=$(get_app_dockerfile "$APP_NAME")
ENTRYPOINT=$(get_app_entrypoint "$APP_NAME" "$LANG")

# Build settings (per app)
GO_VERSION=$(get_app_go_version "$APP_NAME")
GOOS=$(get_app_goos "$APP_NAME")
GOARCH=$(get_app_goarch "$APP_NAME")
CGO_ENABLED=$(get_app_cgo_enabled "$APP_NAME")

BASE_DIR="$GOKKU_BASE_DIR"
# Extract deploy user from git remote or use current user
DEPLOY_USER=$(git remote get-url origin 2>/dev/null | sed 's/@.*//' || echo "${USER:-ubuntu}")
REPO_DIR="$BASE_DIR/repos/$APP_NAME.git"
APP_DIR="$BASE_DIR/apps/$APP_NAME/$ENVIRONMENT"
SERVICE_NAME="$APP_NAME-$ENVIRONMENT"

echo "==> Setting up git-push deployment"
echo "    Project: $GOKKU_PROJECT_NAME"
echo "    App: $APP_NAME"
echo "    Environment: $ENVIRONMENT"
echo "    Build Type: $BUILD_TYPE"
echo "    Language: $LANG"
echo "    Service: $SERVICE_NAME"
if [ "$BUILD_TYPE" = "systemd" ]; then
    echo "    Binary: $BINARY_NAME"
else
    echo "    Image: $APP_NAME:latest"
fi
echo ""

# Create directory structure
sudo mkdir -p $REPO_DIR
sudo mkdir -p $APP_DIR/{releases,shared}
sudo chown -R $DEPLOY_USER:$DEPLOY_USER $BASE_DIR

# Setup Git namespace for short paths
# This allows using: git remote add prod ubuntu@server:api
# Instead of: git remote add prod ubuntu@server:api
USER_HOME=$(eval echo ~$DEPLOY_USER)

# Create Git namespace directory if it doesn't exist
if [ ! -d "$USER_HOME/.git-namespace" ]; then
    sudo -u $DEPLOY_USER mkdir -p $USER_HOME/.git-namespace
fi

# Create symlink for each app (allows short names)
if [ ! -L "$USER_HOME/$APP_NAME.git" ]; then
    sudo -u $DEPLOY_USER ln -sf $REPO_DIR $USER_HOME/$APP_NAME.git
    echo "==> Created Git shortcut: $USER_HOME/$APP_NAME.git -> $REPO_DIR"
fi

# Initialize bare git repository (only if doesn't exist)
if [ ! -d "$REPO_DIR/refs" ]; then
    cd $REPO_DIR
    git init --bare
    echo "==> Git repository initialized at $REPO_DIR"
else
    echo "==> Git repository already exists at $REPO_DIR"
fi

# Create post-receive hook for this environment based on build type
if [ "$BUILD_TYPE" = "docker" ]; then
    HOOK_TEMPLATE="$SCRIPT_DIR/hooks/post-receive-docker.template"
else
    HOOK_TEMPLATE="$SCRIPT_DIR/hooks/post-receive-systemd.template"
fi

if [ ! -f "$HOOK_TEMPLATE" ]; then
    echo "Error: Hook template not found: $HOOK_TEMPLATE"
    exit 1
fi

# Copy template to hook file
cp "$HOOK_TEMPLATE" "$REPO_DIR/hooks/post-receive-$ENVIRONMENT"

# Get Docker base image if needed
if [ "$BUILD_TYPE" = "docker" ]; then
    BASE_IMAGE=$(get_app_base_image "$APP_NAME" "$LANG")
    if [ -z "$BASE_IMAGE" ]; then
        case "$LANG" in
            go) BASE_IMAGE="golang:${GO_VERSION}-alpine" ;;
            python) BASE_IMAGE="python:3.11-slim" ;;
            *) BASE_IMAGE="alpine:latest" ;;
        esac
    fi
fi

# Replace placeholders
sed -i.bak \
    -e "s|__APP_NAME__|$APP_NAME|g" \
    -e "s|__ENVIRONMENT__|$ENVIRONMENT|g" \
    -e "s|__BASE_DIR__|$BASE_DIR|g" \
    -e "s|__BINARY_NAME__|$BINARY_NAME|g" \
    -e "s|__BUILD_PATH__|$BUILD_PATH|g" \
    -e "s|__BUILD_WORKDIR__|$BUILD_WORKDIR|g" \
    -e "s|__ENV_BRANCH__|${ENV_BRANCH:-main}|g" \
    -e "s|__GOOS__|$GOOS|g" \
    -e "s|__GOARCH__|$GOARCH|g" \
    -e "s|__CGO_ENABLED__|$CGO_ENABLED|g" \
    -e "s|__KEEP_RELEASES__|$(get_app_keep_releases $APP_NAME)|g" \
    -e "s|__LANG__|$LANG|g" \
    -e "s|__DOCKERFILE__|${DOCKERFILE}|g" \
    -e "s|__ENTRYPOINT__|${ENTRYPOINT}|g" \
    -e "s|__BASE_IMAGE__|${BASE_IMAGE}|g" \
    -e "s|__KEEP_IMAGES__|$(get_app_keep_images $APP_NAME)|g" \
    $REPO_DIR/hooks/post-receive-$ENVIRONMENT

rm -f $REPO_DIR/hooks/post-receive-$ENVIRONMENT.bak
chmod +x $REPO_DIR/hooks/post-receive-$ENVIRONMENT

# Create main post-receive that routes to environment-specific hooks
cat > $REPO_DIR/hooks/post-receive << 'MAIN_HOOK_EOF'
#!/bin/bash
# Router hook - calls environment-specific hooks based on branch

while read oldrev newrev refname; do
    branch=$(git rev-parse --symbolic --abbrev-ref $refname)

    echo "==> Received push to branch: $branch"

    # Try to find matching environment for this branch
    # Check all environment-specific hooks
    HOOK_FOUND=false

    for hook in $(ls -1 hooks/post-receive-* 2>/dev/null | grep -v "\.bak$"); do
        ENV=$(basename $hook | sed 's/post-receive-//')

        # Check if this environment's branch matches
        # This is a simple match; could be improved with config lookup
        if [[ "$branch" == "main" ]] || [[ "$branch" == "master" ]]; then
            if [[ "$ENV" == "production" ]]; then
                echo "==> Deploying to $ENV environment..."
                $hook
                HOOK_FOUND=true
                break
            fi
        elif [[ "$branch" == "$ENV" ]] || [[ "$branch" == "staging" && "$ENV" == "staging" ]]; then
            echo "==> Deploying to $ENV environment..."
            $hook
            HOOK_FOUND=true
            break
        fi
    done

    if [ "$HOOK_FOUND" = false ]; then
        echo "Warning: Branch '$branch' not mapped to any environment"
        echo "Available environments:"
        ls -1 hooks/post-receive-* 2>/dev/null | grep -v "\.bak$" | sed 's/.*post-receive-/  - /'
        echo ""
        echo "To deploy, push to one of the configured branches"
    fi
done
MAIN_HOOK_EOF

chmod +x $REPO_DIR/hooks/post-receive

# Create systemd service based on build type
if [ "$BUILD_TYPE" = "docker" ]; then
    # Docker-based service
    sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << SERVICE_EOF
[Unit]
Description=$APP_NAME ($ENVIRONMENT) - Docker
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=$DEPLOY_USER
Restart=$(get_app_restart_policy $APP_NAME)
RestartSec=$(get_app_restart_delay $APP_NAME)

# Stop old container
ExecStartPre=-/usr/bin/docker stop $SERVICE_NAME
ExecStartPre=-/usr/bin/docker rm $SERVICE_NAME

# Start new container
ExecStart=/usr/bin/docker run --rm --name $SERVICE_NAME \\
  --env-file $APP_DIR/shared/.env \\
  -p \${PORT}:\${PORT} \\
  $APP_NAME:latest

# Cleanup
ExecStop=/usr/bin/docker stop $SERVICE_NAME

# Load env vars (for PORT mapping)
EnvironmentFile=$APP_DIR/shared/.env

[Install]
WantedBy=multi-user.target
SERVICE_EOF
else
    # Systemd-based service (binary)
    sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << SERVICE_EOF
[Unit]
Description=$APP_NAME ($ENVIRONMENT)
After=network.target

[Service]
Type=simple
User=$DEPLOY_USER
WorkingDirectory=$APP_DIR/current
ExecStart=$APP_DIR/current/$BINARY_NAME
Restart=$(get_app_restart_policy $APP_NAME)
RestartSec=$(get_app_restart_delay $APP_NAME)

# Load env vars from shared .env file
EnvironmentFile=$APP_DIR/shared/.env

# Security
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
SERVICE_EOF
fi

sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME

# Create initial .env if doesn't exist
if [ ! -f "$APP_DIR/shared/.env" ]; then
    cat > $APP_DIR/shared/.env << ENV_EOF
# Environment: $ENVIRONMENT
# App: $APP_NAME
# Generated: $(date)

# Add your environment variables here
# Example:
# API_KEY=your-key-here
# DATABASE_URL=postgres://...
# PORT=8080
ENV_EOF
    echo "==> Created initial .env file at $APP_DIR/shared/.env"
fi

echo ""
echo "==> Setup complete for $APP_NAME ($ENVIRONMENT)!"
echo ""
echo "Build Type: $BUILD_TYPE"
echo "Language: $LANG"
echo ""
echo "Git remote (add to your local repo):"
echo "  git remote add $APP_NAME-$ENVIRONMENT $DEPLOY_USER@YOUR_EC2_HOST:$APP_NAME"
echo ""
echo "Deploy:"
echo "  git push $APP_NAME-$ENVIRONMENT ${ENV_BRANCH:-main}"
echo ""
echo "Manage (from your machine):"
echo "  gokku config set KEY=VALUE --remote $APP_NAME-$ENVIRONMENT"
echo "  gokku restart $APP_NAME $ENVIRONMENT --remote $APP_NAME-$ENVIRONMENT"
echo "  gokku logs $APP_NAME $ENVIRONMENT -f --remote $APP_NAME-$ENVIRONMENT"
echo ""
echo "On server:"
echo "  sudo systemctl status $SERVICE_NAME"
echo "  sudo journalctl -u $SERVICE_NAME -f"

if [ "$BUILD_TYPE" = "docker" ]; then
    echo ""
    echo "Docker commands:"
    echo "  Images:     docker images $APP_NAME"
    echo "  Logs:       docker logs $SERVICE_NAME -f"
    echo "  Inspect:    docker inspect $SERVICE_NAME"
    echo "  Exec:       docker exec -it $SERVICE_NAME sh"
fi

