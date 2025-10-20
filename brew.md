# Homebrew Packaging Guide

## Overview

This guide explains how to publish `cwlogs` as a Homebrew package for macOS users. It assumes you have the source in `cwlogs/` and want to distribute signed binaries through a custom tap.

## Prerequisites

- **Homebrew installed** on macOS.
- **GitHub repository** hosting the `cwlogs` source (e.g., `github.com/your-org/cwlogs`).
- **Go toolchain** to produce macOS binaries.
- **Tap repository** where the Homebrew formula will live (e.g., `github.com/your-org/homebrew-cwlogs`).

## Step 1: Prepare a Release Artifact

1. **Build a macOS binary**:
   ```bash
   GOOS=darwin GOARCH=amd64 go build -o cwlogs
   ```
   Repeat with `GOARCH=arm64` if you intend to ship a universal release.
2. **Archive the binary**:
   ```bash
   tar -czf cwlogs-darwin-amd64.tar.gz cwlogs
   ```
3. **Create a GitHub release** for the project tag (e.g., `v1.0.0`) and upload the tarball.
4. **Record the SHA256 checksum** of the archive for the formula:
   ```bash
   shasum -a 256 cwlogs-darwin-amd64.tar.gz
   ```

## Step 2: Create or Update the Tap Repository

1. **Create the tap repo** `homebrew-cwlogs` (public) on GitHub if it does not exist.
2. **Clone the tap locally**:
   ```bash
   git clone git@github.com:your-org/homebrew-cwlogs.git
   cd homebrew-cwlogs
   ```

## Step 3: Write the Formula

Create `Formula/cwlogs.rb` inside the tap repository:

```ruby
class Cwlogs < Formula
  desc "CloudWatch Log Viewer"
  homepage "https://github.com/your-org/cwlogs"
  url "https://github.com/your-org/cwlogs/releases/download/v1.0.0/cwlogs-darwin-amd64.tar.gz"
  sha256 "<SHA256_FROM_STEP_1>"
  version "1.0.0"
  license "MIT"

  def install
    bin.install "cwlogs"
  end

  test do
    assert_match "CloudWatch", shell_output("#{bin}/cwlogs --help", 2)
  end
end
```

Adjust the URL, SHA256, license, and version for each release. If you ship separate builds per architecture, create conditional logic or separate bottles accordingly.

## Step 4: Test the Formula Locally

1. **Run brew install from the tap directory**:
   ```bash
   brew install --build-from-source ./Formula/cwlogs.rb
   ```
2. **Verify execution**:
   ```bash
   cwlogs --help
   ```
3. **Run the Homebrew audit**:
   ```bash
   brew audit --new-formula ./Formula/cwlogs.rb
   ```

## Step 5: Publish the Tap

1. **Commit and push** the new formula:
   ```bash
   git add Formula/cwlogs.rb
   git commit -m "cwlogs 1.0.0 formula"
   git push origin main
   ```
2. Users can now install using:
   ```bash
   brew tap your-org/cwlogs
   brew install cwlogs
   ```

## Step 6: Updating Releases

For each new version:

- **Build and upload** a new release archive tagged with the version.
- **Update the formula** URL, `sha256`, and `version` fields.
- **Retest** and publish the formula.
- Optionally, **add a `bottle do` block** if you generate bottles for arm64 and amd64 to ship precompiled binaries.

## Notes

- Homebrew expects the binary inside the tarball to have executable permissions.
- To support both Intel and Apple Silicon from a single release, provide bottles or universal binaries and adjust the formula accordingly.
- Keep the tap repository clean—one formula per file inside `Formula/`.
