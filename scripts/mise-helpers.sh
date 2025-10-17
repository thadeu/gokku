#!/bin/bash
# Mise helper functions for Gokku deployment

# Check if mise is installed
is_mise_installed() {
    command -v mise >/dev/null 2>&1
}

# Install mise if not present
install_mise() {
    if is_mise_installed; then
        echo "mise is already installed"
        return 0
    fi

    echo "Installing mise..."
    curl https://mise.run | sh

    # Add to PATH for current session
    export PATH="$HOME/.local/bin:$PATH"

    # Verify installation
    if is_mise_installed; then
        echo "mise installed successfully"
        return 0
    else
        echo "ERROR: Failed to install mise"
        return 1
    fi
}

# Check if .tool-versions exists
has_tool_versions() {
    local dir=$1
    [ -f "$dir/.tool-versions" ]
}

# Install mise plugins from config
install_mise_plugins() {
    local plugins_json=$1

    if [ -z "$plugins_json" ]; then
        return 0
    fi

    echo "-----> Installing mise plugins..."

    # Parse JSON and install each plugin
    # Format: name1:url1,name2:url2
    IFS=',' read -ra PLUGINS <<< "$plugins_json"
    for plugin in "${PLUGINS[@]}"; do
        IFS=':' read -r name url <<< "$plugin"
        if [ -n "$name" ] && [ -n "$url" ]; then
            echo "       Installing plugin: $name"
            if mise plugins list | grep -q "^${name}$"; then
                echo "       Plugin $name already installed"
            else
                mise plugins install "$name" "$url" || {
                    echo "ERROR: Failed to install plugin $name"
                    return 1
                }
            fi
        fi
    done

    echo "-----> Plugins installed"
}

# Run mise install
run_mise_install() {
    local dir=$1

    if ! has_tool_versions "$dir"; then
        return 0
    fi

    echo "-----> Detected .tool-versions"

    cd "$dir" || return 1

    # Show what will be installed
    echo "       Tools to install:"
    while IFS= read -r line; do
        # Skip comments and empty lines
        [[ "$line" =~ ^#.*$ ]] && continue
        [[ -z "$line" ]] && continue
        echo "       - $line"
    done < .tool-versions

    # Install tools
    echo "-----> Running mise install..."
    if ! mise install; then
        echo "ERROR: mise install failed"
        return 1
    fi

    echo "-----> Tools installed successfully"
    return 0
}

# Activate mise in current shell
activate_mise() {
    eval "$(mise activate bash)"
}

# Get mise shims path
get_mise_shims_path() {
    echo "$HOME/.local/share/mise/shims"
}

# Check if specific tool is available via mise
has_mise_tool() {
    local tool=$1
    mise list | grep -q "^${tool}"
}

# Generate Dockerfile with mise support
generate_dockerfile_with_mise() {
    local lang=$1
    local base_image=$2
    local entrypoint=$3
    local plugins=$4
    local output=$5

    cat > "$output" << 'DOCKERFILE_MISE'
# Generated Dockerfile with mise support
FROM ubuntu:22.04

# Install dependencies (including build tools for mise)
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    ca-certificates \
    pkg-config \
    autoconf \
    automake \
    libtool \
    yasm \
    nasm \
    libssl-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Install mise
RUN curl https://mise.run | sh
ENV PATH="/root/.local/bin:/root/.local/share/mise/shims:${PATH}"

WORKDIR /app

# Copy .tool-versions first (for cache optimization)
COPY .tool-versions* ./

# Install mise plugins if needed
__PLUGINS_INSTALL__

# Install tools from .tool-versions
RUN if [ -f .tool-versions ]; then \
        mise install; \
    fi

# Copy application code
COPY . .

# Expose port
EXPOSE ${PORT:-8080}

# Run application
CMD ["__ENTRYPOINT__"]
DOCKERFILE_MISE

    # Replace plugins section
    if [ -n "$plugins" ]; then
        local plugins_cmd="RUN "
        IFS=',' read -ra PLUGINS <<< "$plugins"
        for plugin in "${PLUGINS[@]}"; do
            IFS=':' read -r name url <<< "$plugin"
            plugins_cmd+="mise plugins install $name $url && "
        done
        plugins_cmd="${plugins_cmd% && }"

        sed -i "s|__PLUGINS_INSTALL__|$plugins_cmd|g" "$output"
    else
        sed -i "s|__PLUGINS_INSTALL__|# No extra plugins|g" "$output"
    fi

    # Replace entrypoint based on language
    case "$lang" in
        python)
            sed -i "s|__ENTRYPOINT__|python $entrypoint|g" "$output"
            ;;
        nodejs)
            sed -i "s|__ENTRYPOINT__|node $entrypoint|g" "$output"
            ;;
        ruby)
            sed -i "s|__ENTRYPOINT__|ruby $entrypoint|g" "$output"
            ;;
        *)
            sed -i "s|__ENTRYPOINT__|./$entrypoint|g" "$output"
            ;;
    esac
}

