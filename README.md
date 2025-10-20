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

### Prerequisites

- Go 1.19 or later
- AWS CLI configured with appropriate permissions
- Access to CloudWatch Logs

### Build from Source

```bash
git clone <repository-url>
cd cwlogs
go build -o cwlogs
```

## Usage

### Basic Usage

```bash
./cwlogs
```

The application will:
1. Show available AWS profiles
2. Let you select a CloudWatch log group
3. Start streaming logs in real-time

### Controls

#### Navigation
- `↑/↓` or `j/k` - Scroll up/down one line
- `Page Up/Page Down` - Fast scroll
- `g` - Go to top
- `G` or `End` - Go to bottom (enables follow mode)

#### Search
- `/` - Start search
- `Enter` - Execute search
- `n` - Next match
- `N` - Previous match
- `Esc` - Clear search

#### Display Options
- `J` - Toggle between Raw and Formatted modes
- `F` - Toggle follow mode (auto-scroll to new logs)
- `H` - Load more history

#### Other
- `q` - Quit application

### Display Modes

#### Raw Mode
Shows original CloudWatch log messages exactly as received:
```
[12:34:56] 2023/10/15 12:34:56 [INFO] User login successful
[12:34:57] 2023/10/15 12:34:57 [ERROR] Database connection failed
```

#### Formatted Mode
Parses and colorizes structured logs:
```
[12:34:56] GET /api/users 200 2.3ms - Mozilla/5.0...
[12:34:57] POST /api/login 401 1.1ms - Invalid credentials
```

### Search Features

- **Case-insensitive** - Searches ignore case by default
- **Regex support** - Use regular expressions for complex patterns
- **Dual-mode search** - Works in both raw and formatted modes
- **Highlight preservation** - Search highlights remain visible when navigating
- **Auto-centering** - Found matches automatically center in viewport

### Follow Mode

When **Follow Mode** is enabled (press `F`):
- New logs automatically appear at the bottom
- Cursor tracks the latest log entry
- Scrolling up automatically disables follow mode
- Press `G` or `End` to re-enable follow mode

## Configuration

### AWS Setup

Ensure your AWS credentials are configured:

```bash
# Using AWS CLI
aws configure

# Or using AWS SSO
aws sso login --profile your-profile
```

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

### Search Best Practices

- Use specific terms to narrow results quickly
- Search works on both timestamp and log content
- Use `n/N` to navigate through multiple matches
- Clear search (`Esc`) to return to normal browsing

### Performance Tips

- The viewer maintains only the last 5000 logs in memory
- Older logs are automatically removed (no manual cleanup needed)
- Format toggle is instant - switch freely between modes
- Search is cached - repeated searches are fast

## Troubleshooting

### Common Issues

**"No log groups found"**
- Check AWS credentials and permissions
- Verify you have access to CloudWatch Logs in the selected region

**"AWS SSO session expired"**
```bash
aws sso login --profile <your-profile>
```

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

## Development

For developers interested in contributing or understanding the internals, see [DEVELOPER.md](DEVELOPER.md) for detailed architecture documentation.

## License

[Add your license here]
