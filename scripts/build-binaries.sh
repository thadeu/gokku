#!/bin/bash
set -e

echo "Building Gokku binaries for multiple platforms..."

if [ -f "bin/.gitkeep" ]; then
    mv bin/.gitkeep /tmp/.gitkeep
fi
rm -rf bin/
mkdir -p bin/
if [ -f "/tmp/.gitkeep" ]; then
    mv /tmp/.gitkeep bin/.gitkeep
fi

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

    GOOS=$os GOARCH=$arch go build -o "bin/$binary_name" ./cmd/cli

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
