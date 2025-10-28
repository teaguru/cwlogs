# Automated Release & Deployment Guide

## Overview

This guide explains the automated release process for `cwlogs` using GoReleaser. The system automatically builds binaries, creates GitHub releases, and updates the Homebrew tap when version tags are pushed.

## Prerequisites

- **GitHub repository** hosting the `cwlogs` source (`github.com/teaguru/cwlogs`)
- **Homebrew tap repository** (`github.com/teaguru/homebrew-cwlogs`) - created automatically
- **GitHub Actions** enabled on the repository
- **GoReleaser** installed locally for testing (optional)

## Setup Checklist

Before your first release, ensure:

1. ✅ **GitHub repository** `teaguru/cwlogs` exists and has the code
2. ✅ **Homebrew tap repository** `teaguru/homebrew-cwlogs` created as **public**
3. ✅ **GitHub Actions** enabled on the main repository
4. ✅ **Configuration files** committed:
   - `.goreleaser.yml`
   - `.github/workflows/release.yml`
   - Updated `Makefile`

## Automated Release Process

### Quick Release (Recommended)

1. **Ensure all changes are committed and pushed**
2. **Create and push a version tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. **GitHub Actions automatically**:
   - Runs tests
   - Builds binaries for macOS (amd64/arm64) and Linux (amd64/arm64)
   - Creates GitHub release with assets
   - Updates Homebrew tap at `teaguru/homebrew-cwlogs`

### What Gets Created Automatically

When you push a version tag, the system creates:

- **GitHub Release** with:
  - `cwlogs-v1.0.0-Darwin-x86_64.tar.gz` (macOS Intel)
  - `cwlogs-v1.0.0-Darwin-arm64.tar.gz` (macOS Apple Silicon)
  - `cwlogs-v1.0.0-Linux-x86_64.tar.gz` (Linux amd64)
  - `cwlogs-v1.0.0-Linux-arm64.tar.gz` (Linux arm64)
  - `checksums.txt` with SHA256 hashes

- **Homebrew Formula** at `teaguru/homebrew-cwlogs`:
  - Automatically generated `Formula/cwlogs.rb`
  - Correct SHA256 values for both architectures
  - Proper version and download URLs

### Installation for Users

Once released, users can install via:

**Homebrew (macOS):**
```bash
brew tap teaguru/cwlogs
brew install cwlogs
```

**Linux Package Managers:**
```bash
# Debian/Ubuntu
sudo dpkg -i cwlogs_1.0.0_linux_amd64.deb

# RedHat/CentOS/Fedora  
sudo rpm -i cwlogs_1.0.0_linux_amd64.rpm

# Arch Linux
sudo pacman -U cwlogs_1.0.0_linux_amd64.pkg.tar.zst

# Alpine Linux
sudo apk add --allow-untrusted cwlogs_1.0.0_linux_amd64.apk
```

**Verify installation:**
```bash
cwlogs --help
```

## Local Testing & Development

### Test GoReleaser Configuration

Before creating a real release, test the configuration:

```bash
# Test the configuration (dry run)
make release-dry-run

# Create a local snapshot build
make release-snapshot

# Build locally without publishing
make release-local
```

### Manual Release (Fallback)

If you need to create a release manually:

```bash
# Build release artifacts
VERSION=v1.0.0 make release

# Upload to GitHub manually and update Homebrew tap
```

## Configuration Files

The automated release system uses these configuration files:

### `.goreleaser.yml`
- Defines build targets (macOS/Linux, amd64/arm64)
- Configures archive format and naming
- Sets up Homebrew tap integration
- Manages GitHub release creation

### `.github/workflows/release.yml`
- Triggers on version tags (`v*`)
- Runs tests before release
- Executes GoReleaser with proper permissions

### `Makefile` (Updated)
- Added GoReleaser integration targets
- Local testing commands
- Maintains backward compatibility

## Troubleshooting

### Release Failed
- Check GitHub Actions logs in the repository
- Verify the tag format matches `v*` (e.g., `v1.0.0`)
- Ensure tests pass locally: `make test`

### Homebrew Formula Issues
- GoReleaser automatically handles SHA256 calculation
- Formula is generated automatically - no manual editing needed
- **Important**: Create the tap repository `teaguru/homebrew-cwlogs` as **public** on GitHub before first release

### Local Testing
```bash
# Install GoReleaser
brew install goreleaser/tap/goreleaser

# Test configuration
make release-dry-run

# Check what would be built
make release-local
```

## Security Notes

- **No signing configured by default** - can be added later if needed
- **GitHub token** used automatically by Actions
- **Tap repository** must be public for Homebrew to access
- **Release artifacts** are public on GitHub releases

## Version Management

- Use semantic versioning: `v1.0.0`, `v1.0.1`, `v1.1.0`
- Tags trigger releases automatically
- Version is embedded in binary via ldflags
- Changelog generated automatically from git commits
