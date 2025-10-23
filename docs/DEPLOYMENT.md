# Homebrew Packaging Guide

## Overview

This guide explains how to publish `cwlogs` as a Homebrew package for macOS users. It assumes the project lives at `github.com/teaguru/cwlogs` and that you want to distribute signed binaries through a dedicated Homebrew tap.

## Prerequisites

- **Homebrew installed** on macOS.
- **GitHub repository** hosting the `cwlogs` source (`github.com/teaguru/cwlogs`).
- **Go toolchain** to produce macOS binaries.
- **Tap repository** where the Homebrew formula will live (`github.com/teaguru/homebrew-cwlogs`).

## Step 1: Prepare a Release Artefact

1. **Build release artefacts** using the Makefile:
   ```bash
   VERSION=v1.0.0 make release
   ```
   This generates cross-platform binaries, tarballs, and checksums in `dist/`.

2. **Create a GitHub release** at `https://github.com/teaguru/cwlogs/releases` tagged as `v1.0.0`.

3. **Upload artefacts** from `dist/` to the release:
   - `cwlogs-v1.0.0-darwin-amd64.tar.gz`
   - `cwlogs-v1.0.0-darwin-arm64.tar.gz`
   - `checksums.txt`

4. **Note the SHA256 values** from `checksums.txt` for the formula.

## Step 2: Create or Update the Tap Repository

1. **Create the tap repo** `teaguru/homebrew-cwlogs` (public) on GitHub if it does not exist.
2. **Clone the tap locally**:
   ```bash
   git clone git@github.com:teaguru/homebrew-cwlogs.git
   cd homebrew-cwlogs
   ```

## Step 3: Write the Formula

Create `Formula/cwlogs.rb` inside the tap repository:

```ruby
class Cwlogs < Formula
  desc "CloudWatch Log Viewer"
  homepage "https://github.com/teaguru/cwlogs"
  version "1.0.0"

  on_macos do
    arch = Hardware::CPU.arm? ? "arm64" : "amd64"
    url "https://github.com/teaguru/cwlogs/releases/download/v#{version}/cwlogs-#{version}-darwin-#{arch}.tar.gz"
    sha256 "REPLACE_WITH_SHA256_FOR_#{arch.upcase}"
  end

  license "Apache-2.0"

  def install
    binary = Hardware::CPU.arm? ? "cwlogs-arm64" : "cwlogs-amd64"
    bin.install binary => "cwlogs"
  end

  test do
    assert_match "CloudWatch", shell_output("#{bin}/cwlogs --help", 2)
  end
end
```

Replace `version "1.0.0"` and the `sha256` placeholders with the values that match the release you published in Step 1. If you only ship a single universal binary, simplify the `on_macos` block to a single `url` and `sha256`.

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
   brew tap teaguru/cwlogs
   brew install cwlogs
   ```

## Step 6: Updating Releases

For each new version:

1. Run `VERSION=vX.Y.Z make release` to build artefacts.
2. Create GitHub release and upload from `dist/`.
3. Update formula `version` and `sha256` values from `checksums.txt`.
4. Test with `brew install` and `brew audit`.
5. Commit and push the updated formula.

## Notes

- Homebrew expects the binary inside the tarball to have executable permissions.
- To support both Intel and Apple Silicon from a single release, provide bottles or universal binaries and adjust the formula accordingly.
- Keep the tap repository clean—one formula per file inside `Formula/`.
