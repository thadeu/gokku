#!/bin/bash
# Setup Git shortcuts for simplified remote URLs
# This allows using: git remote add prod ubuntu@server:api
# Instead of: git remote add prod ubuntu@server:/opt/gokku/repos/api.git

set -e

BASE_DIR="${GOKKU_BASE_DIR:-/opt/gokku}"
DEPLOY_USER="${DEPLOY_USER:-ubuntu}"

echo "==> Setting up Git shortcuts for Gokku"
echo "    Base directory: $BASE_DIR"
echo "    Deploy user: $DEPLOY_USER"
echo ""

# Get deploy user's home directory
USER_HOME=$(eval echo ~$DEPLOY_USER)

echo "==> Creating Git namespace configuration..."

# Create a symbolic link in user's home directory
# This allows git to resolve short paths like "api" to "/opt/gokku/repos/api.git"
if [ ! -L "$USER_HOME/git" ]; then
    sudo -u $DEPLOY_USER ln -sf $BASE_DIR/repos $USER_HOME/git
    echo "✓ Created symlink: $USER_HOME/git -> $BASE_DIR/repos"
fi

# Configure Git to use the namespace
# This tells git-receive-pack where to find repositories
sudo -u $DEPLOY_USER git config --global gokku.repoPath "$BASE_DIR/repos"

echo ""
echo "==> Git shortcuts configured!"
echo ""
echo "Now you can use simplified remote URLs:"
echo ""
echo "  Instead of:"
echo "    git remote add production $DEPLOY_USER@server:/opt/gokku/repos/api.git"
echo ""
echo "  Use:"
echo "    git remote add production $DEPLOY_USER@server:git/api.git"
echo "    git remote add production $DEPLOY_USER@server:~/git/api.git"
echo ""
echo "Or even simpler with direct path (if repos are in home):"
echo "    git remote add production $DEPLOY_USER@server:api.git"
echo ""

# Create helper script to list available apps
cat > $BASE_DIR/scripts/list-apps.sh << 'EOFSCRIPT'
#!/bin/bash
BASE_DIR="${GOKKU_BASE_DIR:-/opt/gokku}"
echo "Available Gokku apps:"
for repo in $BASE_DIR/repos/*.git; do
    if [ -d "$repo" ]; then
        app=$(basename "$repo" .git)
        echo "  - $app"
    fi
done
EOFSCRIPT

chmod +x $BASE_DIR/scripts/list-apps.sh

echo "==> Helper script created: $BASE_DIR/scripts/list-apps.sh"
echo ""

