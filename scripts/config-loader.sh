#!/bin/bash
# Configuration loader for gokku deployment
# Reads gokku.yml and exports variables

set -e

CONFIG_FILE="${GOKKU_CONFIG:-gokku.yml}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file $CONFIG_FILE not found"
    exit 1
fi

# Simple YAML parser (supports basic key-value pairs)
parse_yaml() {
    local file=$1
    local prefix=$2
    local s='[[:space:]]*'
    local w='[a-zA-Z0-9_-]*'
    local fs=$(echo @|tr @ '\034')

    sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p" $file |
    awk -F$fs '{
        indent = length($1)/2;
        vname[indent] = $2;
        for (i in vname) {if (i > indent) {delete vname[i]}}
        if (length($3) > 0) {
            vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
            printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
        }
    }'
}

# Load configuration
eval $(parse_yaml "$CONFIG_FILE" "GOKKU_")

# Export common variables
export GOKKU_PROJECT_NAME="${GOKKU_project_name:-gokku}"
export GOKKU_BASE_DIR="${GOKKU_project_base_dir:-/opt/gokku}"
export GOKKU_BUILD_WORKDIR="${GOKKU_build_work_dir:-apps/trunk}"
export GOKKU_KEEP_RELEASES="${GOKKU_deployment_keep_releases:-5}"
export GOKKU_PORT_STRATEGY="${GOKKU_port_strategy:-manual}"
export GOKKU_BASE_PORT="${GOKKU_base_port:-5060}"

# Function to get list of apps from config
get_apps() {
    awk '/^apps:/,/^[^ ]/ {if (/^  - name:/) print $3}' "$CONFIG_FILE"
}

# Function to get list of environments for a specific app
get_app_environments() {
    local app_name=$1
    local envs=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/environments:/) in_env=1
        if (in_env && /- name:/) print \$3
        if (in_env && /^    [a-z]/ && !/- name:/ && !/default_env_vars:/ && !/branch:/) in_env=0
    }" "$CONFIG_FILE")

    # If no environments defined, return default
    if [ -z "$envs" ]; then
        echo "production"
    else
        echo "$envs"
    fi
}

# Function to get build path for an app
get_app_build_path() {
    local app_name=$1
    awk "/^  - name: $app_name/,/^  - name:/ {if (/path:/) print \$2}" "$CONFIG_FILE" | head -1
}

# Function to get binary name for an app
get_app_binary_name() {
    local app_name=$1
    local binary=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/binary_name:/) print \$2}" "$CONFIG_FILE")
    echo "${binary:-$app_name}"
}

# Function to get work_dir for an app
get_app_work_dir() {
    local app_name=$1
    local work_dir=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/work_dir:/) print \$2}" "$CONFIG_FILE")
    echo "${work_dir:-.}"
}

# Function to get branch for environment in a specific app
get_app_env_branch() {
    local app_name=$1
    local env_name=$2
    local branch=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/environments:/) in_envs=1
        if (in_envs && /- name: $env_name/) in_target=1
        if (in_target && /branch:/) { print \$2; exit }
        if (in_target && /- name:/ && !/- name: $env_name/) exit
    }" "$CONFIG_FILE" | sed "s/$env_name//")

    # If no branch defined, use defaults based on environment name
    if [ -z "$branch" ]; then
        case "$env_name" in
            production) branch="main" ;;
            staging) branch="staging" ;;
            development|dev) branch="develop" ;;
            *) branch="main" ;;
        esac
    fi

    echo "$branch"
}

# Function to get build type for an app
get_app_build_type() {
    local app_name=$1
    local build_type=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/type:/) print \$2}" "$CONFIG_FILE" | head -1)
    echo "${build_type:-systemd}"
}

# Function to get Go version for an app
get_app_go_version() {
    local app_name=$1
    local go_version=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/go_version:/) print \$2}" "$CONFIG_FILE" | tr -d '"')
    echo "${go_version:-1.25}"
}

# Function to get GOOS for an app
get_app_goos() {
    local app_name=$1
    local goos=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/goos:/) print \$2}" "$CONFIG_FILE")
    echo "${goos:-linux}"
}

# Function to get GOARCH for an app
get_app_goarch() {
    local app_name=$1
    local goarch=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/goarch:/) print \$2}" "$CONFIG_FILE")
    echo "${goarch:-amd64}"
}

# Function to get CGO_ENABLED for an app
get_app_cgo_enabled() {
    local app_name=$1
    local cgo=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/cgo_enabled:/) print \$2}" "$CONFIG_FILE")
    echo "${cgo:-0}"
}

# Function to get language for an app
get_app_lang() {
    local app_name=$1
    local lang=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/lang:/) print \$2}" "$CONFIG_FILE")
    if [ -z "$lang" ]; then
        # Fallback to default
        lang=$(awk '/^defaults:/,/^[^ ]/ {if (/lang:/) print $2}' "$CONFIG_FILE")
    fi
    echo "${lang:-go}"
}

# Function to get dockerfile path for an app
get_app_dockerfile() {
    local app_name=$1
    awk "/^  - name: $app_name/,/^  - name:/ {if (/dockerfile:/) print \$2}" "$CONFIG_FILE"
}

# Function to get entrypoint for an app (Python, Node.js, etc)
get_app_entrypoint() {
    local app_name=$1
    local lang=$2
    local entrypoint=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/entrypoint:/) print \$2}" "$CONFIG_FILE")

    # If not found, use defaults based on language
    if [ -z "$entrypoint" ]; then
        case "$lang" in
            python) entrypoint="main.py" ;;
            nodejs) entrypoint="index.js" ;;
            ruby) entrypoint="app.rb" ;;
            *) entrypoint="main.py" ;;
        esac
    fi

    echo "$entrypoint"
}

# Function to get Docker base image for an app (from app config or defaults)
get_app_base_image() {
    local app_name=$1
    local lang=$2

    # Try to get from app config first
    local base_image=$(awk "/^  - name: $app_name/,/^  - name:/ {if (/base_image:/) {gsub(/\"/, \"\"); print \$2}}" "$CONFIG_FILE")

    # If not found, get from docker defaults
    if [ -z "$base_image" ]; then
        base_image=$(awk "/^docker:/,/^[^ ]/ {if (/${lang}:/) {gsub(/\"/, \"\"); print \$2}}" "$CONFIG_FILE")
    fi

    # If still not found, use hardcoded defaults
    if [ -z "$base_image" ]; then
        case "$lang" in
            go) base_image="golang:1.25-alpine" ;;
            python) base_image="python:3.11-slim" ;;
            nodejs) base_image="node:20-alpine" ;;
            *) base_image="alpine:latest" ;;
        esac
    fi

    echo "$base_image"
}

# Function to get keep_releases for an app
get_app_keep_releases() {
    local app_name=$1
    local keep=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/deployment:/) in_deploy=1
        if (in_deploy && /keep_releases:/) { print \$2; exit }
    }" "$CONFIG_FILE")
    echo "${keep:-5}"
}

# Function to get keep_images for an app (Docker)
get_app_keep_images() {
    local app_name=$1
    local keep=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/deployment:/) in_deploy=1
        if (in_deploy && /keep_images:/) { print \$2; exit }
    }" "$CONFIG_FILE")
    echo "${keep:-5}"
}

# Function to get restart_policy for an app
get_app_restart_policy() {
    local app_name=$1
    local policy=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/deployment:/) in_deploy=1
        if (in_deploy && /restart_policy:/) { print \$2; exit }
    }" "$CONFIG_FILE")
    echo "${policy:-always}"
}

# Function to get restart_delay for an app
get_app_restart_delay() {
    local app_name=$1
    local delay=$(awk "/^  - name: $app_name/,/^  - name:/ {
        if (/deployment:/) in_deploy=1
        if (in_deploy && /restart_delay:/) { print \$2; exit }
    }" "$CONFIG_FILE")
    echo "${delay:-5}"
}

# Function to check if app has mise plugins configured
has_mise_plugins() {
    local app_name=$1
    awk "/^  - name: $app_name/,/^  - name:/ {
        if (/mise:/) in_mise=1
        if (in_mise && /plugins:/) { print \"true\"; exit }
    }" "$CONFIG_FILE" | grep -q "true"
}

# Function to get mise plugins for an app
get_mise_plugins() {
    local app_name=$1

    # Extract plugins in format: name:url,name:url
    awk "/^  - name: $app_name/,/^  - name:/ {
        if (/mise:/) in_mise=1
        if (in_mise && /plugins:/) in_plugins=1
        if (in_plugins && /- name:/) {
            plugin_name=\$3
            getline
            if (/url:/) {
                plugin_url=\$2
                if (plugin_name && plugin_url) {
                    if (result) result=result\",\"
                    result=result plugin_name\":\"plugin_url
                }
            }
        }
        if (in_plugins && /^      [a-z]/ && !/- name:/ && !/url:/) in_plugins=0
    } END { if (result) print result }" "$CONFIG_FILE"
}

