# Gokku Binaries

This directory contains pre-compiled binaries for Gokku CLI, distributed via the universal installer.

## Binary Naming Convention

Binaries follow the pattern: `gokku-{os}-{arch}`

Supported platforms:
- `gokku-linux-amd64` - Linux x86_64
- `gokku-linux-arm64` - Linux ARM64
- `gokku-darwin-amd64` - macOS Intel
- `gokku-darwin-arm64` - macOS Apple Silicon

## Building Binaries

To build binaries for all platforms, run:

```bash
./scripts/build-binaries.sh
```

This requires Go 1.20+ installed and will create binaries for all supported platforms.

## Release Process

When creating a new release/tag:

1. Ensure all code changes are committed
2. Run `./scripts/build-binaries.sh` to build binaries
3. Optionally compress binaries: `cd bin/ && for f in gokku-*; do gzip "$f"; done`
4. Create checksums: `cd bin/ && sha256sum gokku-* > SHA256SUMS`
5. Commit and push the binaries to the repository

## Installation

The universal installer (`install` script) automatically:
- Detects your OS and architecture
- Downloads the appropriate binary from this directory
- Installs it to `/usr/local/bin/gokku`

No compilation required on the target system!
