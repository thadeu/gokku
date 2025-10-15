#!/bin/bash
# Build script to create binaries for multiple platforms
# Run this when creating releases/tags

set -e

echo "Building Gokku binaries for multiple platforms..."

# Clean previous builds
rm -rf bin/
mkdir -p bin/

# Build for different platforms
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    echo "Building for $os/$arch..."

    binary_name="gokku-$os-$arch"

    if [ "$os" = "darwin" ] && [ "$arch" = "arm64" ]; then
        # Apple Silicon
        GOOS=$os GOARCH=$arch go build -o "bin/$binary_name" ./cmd/cli
    else
        GOOS=$os GOARCH=$arch go build -o "bin/$binary_name" ./cmd/cli
    fi

    echo "✓ Built bin/$binary_name"
done

echo ""
echo "Binaries created in bin/ directory:"

ls -la bin/


echo ""
echo "To compress binaries (optional):"
echo "  cd bin/ && for f in gokku-*; do gzip \"\$f\"; done"
echo ""
echo "To create checksums:"
echo "  cd bin/ && sha256sum gokku-* > SHA256SUMS"
