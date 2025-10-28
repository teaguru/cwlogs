# CloudWatch Log Viewer

A fast, terminal-based viewer for AWS CloudWatch logs with real-time streaming and powerful search capabilities.

## Features

- 🚀 **Real-time log streaming** - See new logs as they arrive
- 🔍 **Fast search** - Find logs instantly with regex support
- 🎨 **Dual display modes** - Switch between raw and formatted views
- 📱 **Responsive UI** - Smooth scrolling and navigation
- 💾 **Memory efficient** - Handles long sessions without memory leaks
- 🔄 **Auto-follow** - Automatically scroll to new logs

## Installation

### Homebrew (macOS) - Recommended

```bash
brew tap teaguru/cwlogs
brew install cwlogs
```

### Direct Download

Download the latest release for your platform from [GitHub Releases](https://github.com/teaguru/cwlogs/releases).

### Build from Source

**Prerequisites:**
- Go 1.19 or later
- AWS CLI configured with appropriate permissions
- Access to CloudWatch Logs
- **For clipboard support on Linux:** `xclip`, `xsel`, or `wl-clipboard` package

```bash
git clone https://github.com/teaguru/cwlogs.git
cd cwlogs
make build
```

## Usage

### Basic Usage

**Interactive mode (default):**
```bash
./cwlogs
```

The application will:
1. Show available AWS profiles (if multiple exist)
2. Let you select a CloudWatch log group
3. Start streaming logs in real-time

**Specify AWS profile:**
```bash
./cwlogs --profile myprofile    # Using flag
./cwlogs myprofile              # Using positional argument (shorter)
```

**Specify AWS region:**
```bash
./cwlogs --profile dev --region us-west-2    # Using flags
./cwlogs dev us-west-2                       # Using positional arguments (shorter)
```

**Show help:**
```bash
./cwlogs --help
```

**Check version:**
```bash
./cwlogs --version
```

### Command-Line Options

| Option/Argument | Description | Example |
|-----------------|-------------|---------|
| `profile` | AWS profile name (positional argument) | `cwlogs production` |
| `region` | AWS region (positional argument) | `cwlogs production us-west-2` |
| `--profile <name>` | Use specific AWS profile (flag alternative) | `--profile production` |
| `--region <name>` | Use specific AWS region (overrides profile default) | `--region us-east-1` |
| `--version` | Show version information | `--version` |
| `--help` | Show help and usage examples | `--help` |

### Makefile Targets

```bash
make build        # Build binary
make test         # Run tests
make lint         # Check code quality
make release      # Build release archives for all platforms
make help         # Show all targets
```

### Controls

#### Navigation
- `↑/↓` or `j/k` - Scroll up/down one line
- `Page Up/Page Down` - Fast scroll
- `g` - Go to top
- `G` or `End` - Go to bottom (enables follow mode)

#### Search
- `/` - Start search
- `Enter` - Execute search (starts at latest/newest match)
- `n` - Next match (backward to older logs)
- `N` - Previous match (forward to newer logs)
- `Esc` - Clear search

#### Display Options
- `J` - Toggle between Raw and Formatted modes
- `F` - Toggle follow mode (auto-scroll to new logs)
- `H` - Load more history

#### Copy Text
- `c` - Copy current log line to clipboard (original unformatted message)
- **Mouse/trackpad** - Select any text and copy with Cmd+C/Ctrl+C

#### Other
- `b` or `Backspace` - Go back to log group selection
- `q` - Quit application

### Log Group Selection Controls

When selecting a log group, you can use:
- `↑/↓` or `j/k` - Navigate through log groups
- **Type** - Start filtering log groups by name (automatic)
- `Enter` - Select log group
- `R` - Change AWS region
- `Esc` - Clear filter (if filtering) or quit
- `q` - Quit application

### Display Modes

#### Raw Mode
Shows original CloudWatch log messages exactly as received:
```
[12:34:56] 2023/10/15 12:34:56 [INFO] User login successful
[12:34:57] 2023/10/15 12:34:57 [ERROR] Database connection failed
```

#### Formatted Mode
Parses and colourises structured logs:
```
[12:34:56] GET /api/users 200 2.3ms - Mozilla/5.0...
[12:34:57] POST /api/login 401 1.1ms - Invalid credentials
```

### Search Features

- **Case-insensitive** - Searches ignore case by default
- **Regex support** - Use regular expressions for complex patterns
- **Dual-mode search** - Works in both raw and formatted modes
- **Highlight preservation** - Search highlights remain visible when navigating
- **Auto-centring** - Found matches automatically centre in viewport

### Follow Mode

When **Follow Mode** is enabled (press `F`):
- New logs automatically appear at the bottom
- Cursor tracks the latest log entry
- Scrolling up automatically disables follow mode
- Press `G` or `End` to re-enable follow mode

## Configuration

### AWS Setup

The application supports multiple AWS credential configurations:

**Option 1: Simple setup (most common)**
```bash
# Configure default credentials
aws configure
# Enter your Access Key ID, Secret Access Key, region, and output format
```

**Option 2: Named profiles**
```bash
# Configure a named profile
aws configure --profile myprofile
```

**Option 3: AWS SSO**
```bash
# Configure SSO profile
aws configure sso --profile mysso
aws sso login --profile mysso
```

**Option 4: Environment variables**
```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_DEFAULT_REGION=us-east-1
```

The application will automatically detect available profiles and let you choose, or use the default profile if only one is configured.

### Required Permissions

Your AWS user/role needs these CloudWatch permissions:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:DescribeLogGroups",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

## Tips & Tricks

### Efficient Log Monitoring

1. **Use Follow Mode** (`F`) for real-time monitoring
2. **Search while following** - Search automatically disables follow mode
3. **Toggle formats** (`J`) to see both raw and parsed views
4. **Load history** (`H`) to see older logs when needed

### Command-Line Efficiency

**Quick access to specific environments:**
```bash
# Development environment (shorter syntax)
./cwlogs dev

# Production monitoring in specific region
./cwlogs prod us-west-2

# Staging logs in different region
./cwlogs staging eu-west-1

# Alternative flag syntax
./cwlogs --profile production --region us-east-1
```

**Automation and scripting:**
```bash
#!/bin/bash
# Monitor production logs in specific region
echo "Starting production log monitoring..."
./cwlogs production us-west-2
```

**Integration with other tools:**
```bash
# Use with AWS SSO in specific region
aws sso login --profile company-admin
./cwlogs company-admin eu-central-1
```

### Search Best Practices

- Use specific terms to narrow results quickly
- Search works on both timestamp and log content
- Use `n/N` to navigate through multiple matches
- Clear search (`Esc`) to return to normal browsing

### Copy and Share Logs

**Method 1: Keyboard shortcut**
- Press `c` to copy the current log line to clipboard
- Works on macOS (pbcopy), Linux (xclip/xsel/wl-copy), and Windows (clip)
- **Always copies the original unformatted CloudWatch message** (even in formatted mode)
- Status message confirms successful copy

**Method 2: Mouse/trackpad selection**
- Select text with mouse or trackpad (drag to highlight)
- Copy with `Cmd+C` (macOS) or `Ctrl+C` (Linux/Windows)
- Works with any text in the terminal, including partial lines
- Copies the formatted display text (what you see on screen)

**Why use keyboard vs mouse?**
- **Keyboard (`c`)**: Gets the original CloudWatch message - perfect for debugging, sharing with team, or pasting into other tools
- **Mouse selection**: Gets the formatted display text - useful for copying specific parts or formatted output

### Performance Tips

- The viewer maintains only the last 5000 logs in memory
- Older logs are automatically removed (no manual cleanup needed)
- Format toggle is instant - switch freely between modes
- Search is cached - repeated searches are fast

## Troubleshooting

### Common Issues

**"No AWS profiles found"**
- Run `aws configure` to set up default credentials
- Check that `~/.aws/credentials` or `~/.aws/config` exists
- Verify AWS CLI is installed: `aws --version`

**"No log groups found"**
- Check AWS credentials and permissions
- Verify you have access to CloudWatch Logs in the selected region
- Try: `aws logs describe-log-groups` to test access

**"AWS SSO session expired"**
```bash
aws sso login --profile <your-profile>
```

**"Failed to load AWS configuration"**
- Check your AWS region is set: `aws configure get region`
- Verify credentials: `aws sts get-caller-identity`
- For environment variables, ensure all required vars are set
- **On EC2**: Ensure the instance has an IAM role with CloudWatch permissions attached

**"No log groups found" (region-specific)**
- Verify you're looking in the correct region: `./cwlogs profile us-west-2`
- Check if logs exist in that region: `aws logs describe-log-groups --region us-west-2`
- CloudWatch logs are region-specific - try different regions

**Logs not updating in real-time**
- Check if Follow Mode is enabled (press `F`)
- Verify CloudWatch is receiving new logs
- Try pressing `G` to jump to latest logs

**Search not working**
- Try clearing search with `Esc` and searching again
- Check if you're in the right display mode (`J` to toggle)
- Verify the search term exists in the visible logs

### Performance

The application is designed for long-running sessions:
- Memory usage stays constant (typically ~50MB)
- Handles thousands of logs without slowdown
- Automatic cleanup of old logs
- Responsive UI even with high log volume

## Documentation

- [Usage Guide](docs/USAGE.md) - Advanced features and workflows
- [AWS Setup](docs/AWS-SETUP.md) - Configure AWS credentials and profiles
- [Developer Guide](docs/DEVELOPER.md) - Architecture and implementation details
- [Testing](docs/TESTING.md) - Test coverage and approach
- [Deployment](docs/DEPLOYMENT.md) - Building and publishing releases

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.
