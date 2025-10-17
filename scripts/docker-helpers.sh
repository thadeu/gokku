#!/bin/bash
# Docker helper functions for Gokku deployment
# Provides all-in-one Docker installation and verification

# Check if docker is installed
is_docker_installed() {
    command -v docker >/dev/null 2>&1
}

# Check if docker daemon is running
is_docker_running() {
    docker ps > /dev/null 2>&1
}

# Detect Linux distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f /etc/lsb-release ]; then
        . /etc/lsb-release
        echo "$DISTRIB_ID" | tr '[:upper:]' '[:lower:]'
    else
        echo "unknown"
    fi
}

# Install Docker on Ubuntu/Debian
install_docker_debian() {
    echo "-----> Installing Docker on Debian/Ubuntu..."

    # Update package lists
    sudo apt-get update -qq

    # Install dependencies
    echo "-----> Installing dependencies..."
    sudo apt-get install -y --no-install-recommends \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release > /dev/null 2>&1

    # Add Docker GPG key
    echo "-----> Adding Docker GPG key..."
    curl -fsSL https://download.docker.com/linux/$(lsb_release -si | tr '[:upper:]' '[:lower:]')/gpg | \
        sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg > /dev/null 2>&1

    # Add Docker repository
    echo "-----> Adding Docker repository..."
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
        https://download.docker.com/linux/$(lsb_release -si | tr '[:upper:]' '[:lower:]') \
        $(lsb_release -cs) stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Update and install Docker
    echo "-----> Installing Docker CE..."
    sudo apt-get update -qq
    sudo apt-get install -y --no-install-recommends \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        docker-compose-plugin > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "-----> Docker installed successfully"
        return 0
    else
        echo "ERROR: Failed to install Docker"
        return 1
    fi
}

# Install Docker on Amazon Linux / RHEL / CentOS
install_docker_rhel() {
    echo "-----> Installing Docker on RHEL/CentOS/Amazon Linux..."

    # Update package lists
    sudo yum update -y -q

    # Install Docker
    echo "-----> Installing Docker..."
    sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "-----> Docker installed successfully"
        return 0
    else
        echo "ERROR: Failed to install Docker"
        return 1
    fi
}

# Install Docker on Alpine
install_docker_alpine() {
    echo "-----> Installing Docker on Alpine..."
    sudo apk add --no-cache docker docker-compose

    if [ $? -eq 0 ]; then
        echo "-----> Docker installed successfully"
        return 0
    else
        echo "ERROR: Failed to install Docker"
        return 1
    fi
}

# Main install function
install_docker() {
    # Check if already installed
    if is_docker_installed; then
        echo "-----> Docker is already installed"

        # Start docker if not running
        if ! is_docker_running; then
            echo "-----> Starting Docker daemon..."
            sudo systemctl start docker

            # Enable docker to start on boot
            sudo systemctl enable docker > /dev/null 2>&1
        fi

        return 0
    fi

    # Detect distribution
    DISTRO=$(detect_distro)
    echo "-----> Detected Linux distribution: $DISTRO"

    case "$DISTRO" in
        ubuntu|debian)
            install_docker_debian
            ;;
        centos|rhel|fedora|amzn)
            install_docker_rhel
            ;;
        alpine)
            install_docker_alpine
            ;;
        *)
            echo "ERROR: Unsupported Linux distribution: $DISTRO"
            echo "Please install Docker manually from https://docs.docker.com/engine/install/"
            return 1
            ;;
    esac

    if [ $? -ne 0 ]; then
        return 1
    fi

    # Start Docker service
    echo "-----> Starting Docker daemon..."
    sudo systemctl start docker || {
        echo "ERROR: Failed to start Docker daemon"
        return 1
    }

    # Enable Docker to start on boot
    echo "-----> Enabling Docker to start on boot..."
    sudo systemctl enable docker > /dev/null 2>&1

    # Add current user to docker group (optional, for easier use)
    if id -nG "$USER" | grep -qw "docker"; then
        echo "-----> User already in docker group"
    else
        echo "-----> Adding user to docker group..."
        sudo usermod -aG docker "$USER" > /dev/null 2>&1
        echo "       Note: You may need to log out and log back in for group changes to take effect"
    fi

    # Wait for Docker daemon
    echo "-----> Waiting for Docker daemon to be ready..."
    local max_attempts=10
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker ps > /dev/null 2>&1; then
            echo "-----> Docker is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    echo "ERROR: Docker daemon failed to start"
    return 1
}

# Verify Docker installation and functionality
verify_docker() {
    echo "-----> Verifying Docker installation..."

    if ! is_docker_installed; then
        echo "ERROR: Docker is not installed"
        return 1
    fi

    DOCKER_VERSION=$(docker --version)
    echo "       Docker version: $DOCKER_VERSION"

    if ! is_docker_running; then
        echo "ERROR: Docker daemon is not running"
        return 1
    fi

    # Test Docker by running hello-world
    echo "-----> Testing Docker with hello-world image..."
    if docker run --rm hello-world > /dev/null 2>&1; then
        echo "-----> Docker is working correctly"
        return 0
    else
        echo "ERROR: Docker test failed"
        return 1
    fi
}

# Get Docker status
get_docker_status() {
    if is_docker_installed; then
        if is_docker_running; then
            echo "Docker: $(docker --version) - running"
            return 0
        else
            echo "Docker: installed but not running"
            return 1
        fi
    else
        echo "Docker: not installed"
        return 1
    fi
}

# Check if zero-downtime deployment is enabled
is_zero_downtime_enabled() {
    local env_file=$1

    if [ ! -f "$env_file" ]; then
        # Default: enabled
        return 0
    fi

    # Extract ZERO_DOWNTIME from env file
    local zero_downtime=$(grep "^ZERO_DOWNTIME=" "$env_file" 2>/dev/null | cut -d= -f2 | tr -d ' ')

    case "$zero_downtime" in
        1|true|True|TRUE|yes|Yes|YES|on|On|ON|y)
            return 0  # Enabled
            ;;
        0|false|False|FALSE|no|No|NO|off|Off|OFF|n)
            return 1  # Disabled
            ;;
        *)
            # Default: enabled if not specified
            return 0
            ;;
    esac
}

# Standard deployment (kill and restart)
standard_deploy() {
    local app_name=$1
    local image_tag=$2
    local env_file=$3
    local release_dir=$4
    local service_name=$5

    local blue_name="${app_name}-blue"

    echo "=====> Starting Standard Deployment"
    echo "-----> Stopping old container: $blue_name"

    # Stop and remove old container
    if docker ps -a --format '{{.Names}}' | grep -q "^${blue_name}$"; then
        docker stop "$blue_name" 2>/dev/null || true
        docker rm "$blue_name" 2>/dev/null || true
        sleep 2
    fi

    # Get Docker configuration from gokku.yml
    local network_mode=$(get_app_docker_network_mode "$app_name")
    local docker_ports=$(get_app_docker_ports "$app_name")

    echo "-----> Network mode: $network_mode"

    # Build docker run command
    local docker_cmd="sudo docker run -d --name $blue_name"

    # Add restart policy
    docker_cmd="$docker_cmd --restart no"

    # Add network mode
    docker_cmd="$docker_cmd --network $network_mode"

    # Add port mappings
    if [ "$network_mode" != "host" ]; then
        if [ -n "$docker_ports" ]; then
            # Use ports from gokku.yml
            while IFS= read -r port_mapping; do
                [ -n "$port_mapping" ] && docker_cmd="$docker_cmd -p $port_mapping"
            done <<< "$docker_ports"
            echo "-----> Using ports from gokku.yml"
        else
            # Fallback to PORT from env
            local container_port=$(get_container_port "$env_file" 8080)
            docker_cmd="$docker_cmd -p $container_port:${container_port}"
            echo "-----> Using port: $container_port"
        fi
    else
        echo "-----> Using host network (all ports exposed)"
    fi

    # Add environment file if exists
    if [ -f "$env_file" ]; then
        docker_cmd="$docker_cmd --env-file $env_file"
    fi

    # Add working directory volume
    docker_cmd="$docker_cmd -v $release_dir:/app"

    # Add image
    docker_cmd="$docker_cmd ${app_name}:${image_tag}"

    echo "-----> Starting new container: $blue_name"

    # Run container
    if ! eval "$docker_cmd"; then
        echo "ERROR: Failed to start container"
        return 1
    fi

    # Wait for container to be ready
    echo "-----> Waiting for container to be ready..."
    sleep 5

    # Check if container is running
    if ! docker ps --format '{{.Names}}' | grep -q "^${blue_name}$"; then
        echo "ERROR: Container failed to start"
        docker logs "$blue_name" 2>&1 | tail -30
        return 1
    fi

    echo "=====> Standard Deployment Complete!"
    echo "-----> Active container: ${blue_name}"
    echo "-----> Running image: $image_tag"
    echo "-----> Port: $container_port"

    return 0
}

# Determine and execute deployment strategy
deploy_container() {
    local app_name=$1
    local image_tag=$2
    local env_file=$3
    local release_dir=$4
    local service_name=$5
    local health_check_timeout=${6:-60}

    if is_zero_downtime_enabled "$env_file"; then
        echo "=====> ZERO_DOWNTIME deployment enabled"
        blue_green_deploy "$app_name" "$image_tag" "$env_file" "$release_dir" "$service_name" "$health_check_timeout"
    else
        echo "=====> ZERO_DOWNTIME deployment disabled"
        standard_deploy "$app_name" "$image_tag" "$env_file" "$release_dir" "$service_name"
    fi
}

# Blue/Green deployment functions for zero-downtime updates

# Start container (green)
start_green_container() {
    local app_name=$1
    local container_port=$2
    local env_file=$3
    local image_tag=$4
    local release_dir=$5
    local service_name=$6

    local green_name="${app_name}-green"

    echo "-----> Starting green container: $green_name"

    # Stop and remove old green container if exists
    if docker ps -a --format '{{.Names}}' | grep -q "^${green_name}$"; then
        echo "       Removing old green container..."
        docker stop "$green_name" 2>/dev/null || true
        docker rm "$green_name" 2>/dev/null || true
    fi

    # Get Docker configuration from gokku.yml
    local network_mode=$(get_app_docker_network_mode "$app_name")
    local docker_ports=$(get_app_docker_ports "$app_name")

    # Build docker run command
    local docker_cmd="sudo docker run -d --name $green_name"

    # Add restart policy
    docker_cmd="$docker_cmd --restart no"

    # Add network mode
    docker_cmd="$docker_cmd --network $network_mode"

    # Add port mappings
    if [ "$network_mode" != "host" ]; then
        if [ -n "$docker_ports" ]; then
            # Use ports from gokku.yml
            while IFS= read -r port_mapping; do
                [ -n "$port_mapping" ] && docker_cmd="$docker_cmd -p $port_mapping"
            done <<< "$docker_ports"
        else
            # Fallback to container_port parameter
            docker_cmd="$docker_cmd -p $container_port:${container_port}"
        fi
    fi

    # Add environment file if exists
    if [ -f "$env_file" ]; then
        docker_cmd="$docker_cmd --env-file $env_file"
    fi

    # Add working directory volume
    docker_cmd="$docker_cmd -v $release_dir:/app"

    # Add image
    docker_cmd="$docker_cmd ${app_name}:${image_tag}"

    # Run container
    if ! eval "$docker_cmd"; then
        echo "ERROR: Failed to start green container"
        return 1
    fi

    echo "-----> Green container started ($green_name)"
    return 0
}

# Wait for green container to be healthy
wait_for_green_health() {
    local app_name=$1
    local max_wait=${2:-60}  # Default 60 seconds

    local green_name="${app_name}-green"
    local start_time=$(date +%s)

    echo "-----> Waiting for green container to be healthy (max ${max_wait}s)..."

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -gt $max_wait ]; then
            echo "ERROR: Green container failed to become healthy within ${max_wait}s"
            return 1
        fi

        # Check container status
        local status=$(docker inspect "$green_name" --format='{{.State.Health.Status}}' 2>/dev/null || echo "")

        case "$status" in
            "healthy")
                echo "-----> Green container is healthy!"
                return 0
                ;;
            "starting")
                echo "       Starting... ($elapsed/${max_wait}s)"
                sleep 2
                ;;
            "unhealthy")
                echo "ERROR: Green container is unhealthy"
                docker logs "$green_name" 2>&1 | tail -20
                return 1
                ;;
            "")
                # No health check, wait a bit and assume ready
                sleep 3
                echo "-----> Green container ready (no health check configured)"
                return 0
                ;;
        esac
    done
}

# Switch traffic from blue to green
switch_traffic_blue_to_green() {
    local app_name=$1
    local container_port=$2

    local blue_name="${app_name}-blue"
    local green_name="${app_name}-green"

    echo "-----> Switching traffic: blue → green"

    # Stop accepting connections on blue
    if docker ps --format '{{.Names}}' | grep -q "^${blue_name}$"; then
        echo "       Pausing blue container..."
        docker pause "$blue_name" 2>/dev/null || true
        sleep 2
    fi

    # Rename containers (atomic swap)
    echo "       Swapping container names..."

    # Temporary rename old blue to blue-old
    if docker ps -a --format '{{.Names}}' | grep -q "^${blue_name}$"; then
        docker rename "$blue_name" "${blue_name}-old" 2>/dev/null || true
    fi

    # Rename green to blue
    docker rename "$green_name" "$blue_name" 2>/dev/null || {
        echo "ERROR: Failed to rename green container to blue"
        return 1
    }

    # Set proper restart policy for new blue container
    docker update --restart always "$blue_name" > /dev/null 2>&1 || true

    echo "-----> Traffic switch complete (green → blue)"
    return 0
}

# Cleanup old blue container
cleanup_old_blue_container() {
    local app_name=$1

    local old_blue_name="${app_name}-blue-old"

    echo "-----> Cleaning up old blue container..."

    if docker ps -a --format '{{.Names}}' | grep -q "^${old_blue_name}$"; then
        # Give it time to drain connections
        echo "       Waiting 5s before removing old container..."
        sleep 5

        echo "       Removing old blue container..."
        docker stop "$old_blue_name" 2>/dev/null || true
        docker rm "$old_blue_name" 2>/dev/null || true

        echo "-----> Old container cleaned up"
    fi
}

# Recreate active container with new environment variables
recreate_active_container() {
    local app_name=$1
    local env_file=$2
    local app_dir=$3

    # Determine which container is active
    local active_container=""
    if sudo docker ps --format '{{.Names}}' | grep -q "^${app_name}-blue$"; then
        active_container="${app_name}-blue"
    elif sudo docker ps --format '{{.Names}}' | grep -q "^${app_name}-green$"; then
        active_container="${app_name}-green"
    else
        echo "ERROR: No active container found for $app_name"
        return 1
    fi

    echo "-----> Recreating container: $active_container"

    # Get current image
    local image=$(sudo docker inspect "$active_container" --format='{{.Config.Image}}' 2>/dev/null)
    if [ -z "$image" ]; then
        echo "ERROR: Could not determine container image"
        return 1
    fi

    echo "       Using image: $image"

    # Get Docker configuration from gokku.yml
    local network_mode=$(get_app_docker_network_mode "$app_name")
    local docker_ports=$(get_app_docker_ports "$app_name")

    echo "       Network mode: $network_mode"

    # Stop and remove old container
    echo "       Stopping old container..."
    sudo docker stop "$active_container" >/dev/null 2>&1 || true
    sudo docker rm "$active_container" >/dev/null 2>&1 || true

    # Build docker run command
    local docker_cmd="sudo docker run -d --name $active_container --restart always --network $network_mode"

    # Add port mappings
    if [ "$network_mode" != "host" ]; then
        if [ -n "$docker_ports" ]; then
            # Use ports from gokku.yml
            while IFS= read -r port_mapping; do
                [ -n "$port_mapping" ] && docker_cmd="$docker_cmd -p $port_mapping"
            done <<< "$docker_ports"
            echo "       Using ports from gokku.yml"
        else
            # Fallback to PORT from env
            local port=$(get_container_port "$env_file" "8080")
            docker_cmd="$docker_cmd -p ${port}:${port}"
            echo "       Using port: $port"
        fi
    else
        echo "       Using host network (all ports exposed)"
    fi

    # Add env file and image
    docker_cmd="$docker_cmd --env-file $env_file $image"

    # Start new container with same name and updated env
    echo "       Starting new container with updated configuration..."
    if ! eval "$docker_cmd >/dev/null"; then
        echo "ERROR: Failed to recreate container"
        return 1
    fi

    if [ $? -eq 0 ]; then
        echo "✓ Container recreated successfully with new environment"
        return 0
    else
        echo "ERROR: Failed to recreate container"
        return 1
    fi
}

# Get container port from env file
get_container_port() {
    local env_file=$1
    local default_port=${2:-8080}

    if [ ! -f "$env_file" ]; then
        echo "$default_port"
        return
    fi

    # Extract PORT from env file
    local port=$(grep "^PORT=" "$env_file" | cut -d= -f2 | tr -d ' ')

    if [ -n "$port" ]; then
        echo "$port"
    else
        echo "$default_port"
    fi
}

# Check if blue container exists and is healthy
has_running_blue_container() {
    local app_name=$1
    local blue_name="${app_name}-blue"

    if docker ps --format '{{.Names}}' | grep -q "^${blue_name}$"; then
        return 0
    fi
    return 1
}

# Perform blue/green deployment
blue_green_deploy() {
    local app_name=$1
    local image_tag=$2
    local env_file=$3
    local release_dir=$4
    local service_name=$5
    local health_check_timeout=${6:-60}

    echo "=====> Starting Blue/Green Deployment"

    # Get port from env file
    local container_port=$(get_container_port "$env_file" 8080)
    echo "-----> Using port: $container_port"

    # Start green container
    start_green_container "$app_name" "$container_port" "$env_file" "$image_tag" "$release_dir" "$service_name" || {
        echo "ERROR: Failed to start green container"
        return 1
    }

    # Wait for green to be healthy
    wait_for_green_health "$app_name" "$health_check_timeout" || {
        echo "ERROR: Green container failed health check"
        docker stop "${app_name}-green" 2>/dev/null || true
        docker rm "${app_name}-green" 2>/dev/null || true
        return 1
    }

    # Check if we have an existing blue container
    if has_running_blue_container "$app_name"; then
        # Switch traffic: blue → green
        switch_traffic_blue_to_green "$app_name" "$container_port" || {
            echo "ERROR: Failed to switch traffic"
            docker stop "${app_name}-green" 2>/dev/null || true
            docker rm "${app_name}-green" 2>/dev/null || true
            return 1
        }

        # Cleanup old blue
        cleanup_old_blue_container "$app_name"
    else
        # First deployment, just rename green to blue
        echo "-----> First deployment, activating green as blue"
        docker rename "${app_name}-green" "${app_name}-blue" 2>/dev/null || {
            echo "ERROR: Failed to rename green to blue"
            return 1
        }
        docker update --restart always "${app_name}-blue" > /dev/null 2>&1 || true
    fi

    echo "=====> Blue/Green Deployment Complete!"
    echo "-----> Active container: ${app_name}-blue"
    echo "-----> Running image: $image_tag"
    echo "-----> Port: $container_port"

    return 0
}

# Rollback to previous blue container
blue_green_rollback() {
    local app_name=$1

    local blue_name="${app_name}-blue"
    local old_blue_name="${app_name}-blue-old"

    echo "=====> Starting Blue/Green Rollback"

    # Check if old blue exists
    if ! docker ps -a --format '{{.Names}}' | grep -q "^${old_blue_name}$"; then
        echo "ERROR: No previous blue container found for rollback"
        return 1
    fi

    echo "-----> Stopping current blue container..."
    docker stop "$blue_name" 2>/dev/null || true

    echo "-----> Restoring previous blue container..."
    docker rename "$old_blue_name" "$blue_name" 2>/dev/null || {
        echo "ERROR: Failed to restore previous blue container"
        return 1
    }

    echo "-----> Starting previous blue container..."
    docker start "$blue_name" 2>/dev/null || {
        echo "ERROR: Failed to start previous blue container"
        return 1
    }

    # Wait for container to be ready
    sleep 5

    echo "=====> Blue/Green Rollback Complete!"
    echo "-----> Active container: $blue_name"

    return 0
}

