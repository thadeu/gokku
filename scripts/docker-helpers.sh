#!/bin/bash
# Docker helper functions for deployment

# Generate Dockerfile from template
generate_dockerfile() {
    local lang=$1
    local app_name=$2
    local build_path=$3
    local entrypoint=$4
    local output_file=$5
    local base_image=$6
    
    local template_file="$SCRIPT_DIR/templates/Dockerfile.${lang}.template"
    
    if [ ! -f "$template_file" ]; then
        echo "ERROR: Template not found: $template_file"
        return 1
    fi
    
    # Read template and replace placeholders
    cat "$template_file" | \
        sed "s|{{APP_NAME}}|$app_name|g" | \
        sed "s|{{BUILD_PATH}}|$build_path|g" | \
        sed "s|{{MAIN_FILE}}|$entrypoint|g" | \
        sed "s|{{GO_BASE_IMAGE}}|$base_image|g" | \
        sed "s|{{PYTHON_BASE_IMAGE}}|$base_image|g" | \
        sed "s|{{CGO_ENABLED}}|${GOKKU_build_cgo_enabled:-0}|g" | \
        sed "s|{{GOOS}}|${GOKKU_build_goos:-linux}|g" | \
        sed "s|{{GOARCH}}|${GOKKU_build_goarch:-amd64}|g" \
        > "$output_file"
    
    echo "Generated Dockerfile at $output_file"
}

# Build Docker image
build_docker_image() {
    local app_name=$1
    local release_dir=$2
    local dockerfile=$3
    local tag=$4
    
    echo "-----> Building Docker image: $app_name:$tag"
    
    cd "$release_dir"
    
    if ! docker build -f "$dockerfile" -t "$app_name:$tag" .; then
        echo "ERROR: Docker build failed"
        return 1
    fi
    
    # Tag as latest
    docker tag "$app_name:$tag" "$app_name:latest"
    
    echo "-----> Image built successfully"
    echo "       Tags: $app_name:$tag, $app_name:latest"
}

# Stop and remove old container
stop_container() {
    local container_name=$1
    
    if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
        echo "-----> Stopping old container: $container_name"
        docker stop "$container_name" 2>/dev/null || true
        docker rm "$container_name" 2>/dev/null || true
    fi
}

# Start Docker container
start_container() {
    local container_name=$1
    local image_name=$2
    local env_file=$3
    local port_mapping=$4
    
    echo "-----> Starting container: $container_name"
    
    # Build docker run command
    local docker_cmd="docker run -d --name $container_name --restart always"
    
    # Add env file if exists
    if [ -f "$env_file" ]; then
        docker_cmd="$docker_cmd --env-file $env_file"
    fi
    
    # Add port mapping if provided
    if [ -n "$port_mapping" ]; then
        docker_cmd="$docker_cmd -p $port_mapping"
    fi
    
    # Add image
    docker_cmd="$docker_cmd $image_name"
    
    # Run container
    if ! eval "$docker_cmd"; then
        echo "ERROR: Failed to start container"
        return 1
    fi
    
    echo "-----> Container started successfully"
}

# Check if container is running
check_container() {
    local container_name=$1
    
    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
        echo "-----> Container is running"
        return 0
    else
        echo "ERROR: Container is not running"
        docker logs "$container_name" 2>&1 | tail -20
        return 1
    fi
}

# Cleanup old Docker images
cleanup_old_images() {
    local app_name=$1
    local keep_count=$2
    
    echo "-----> Cleaning up old images..."
    
    # Get all images for this app (excluding 'latest')
    local images=$(docker images "$app_name" --format "{{.Tag}}" | grep -v "^latest$" | sort -r | tail -n +$((keep_count + 1)))
    
    if [ -n "$images" ]; then
        for tag in $images; do
            echo "       Removing image: $app_name:$tag"
            docker rmi "$app_name:$tag" 2>/dev/null || true
        done
    fi
}

# Get port mapping from env file
get_port_mapping() {
    local env_file=$1
    
    if [ ! -f "$env_file" ]; then
        echo ""
        return
    fi
    
    # Extract PORT from env file
    local port=$(grep "^PORT=" "$env_file" | cut -d= -f2 | tr -d ' ')
    
    if [ -n "$port" ]; then
        echo "${port}:${port}"
    else
        echo ""
    fi
}

