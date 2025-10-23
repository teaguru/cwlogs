# CloudWatch Log Viewer Documentation

## Quick Start
- [Main README](../README.md) - Installation, basic usage, and features

## User Guides
- [AWS Setup](AWS-SETUP.md) - Configure AWS credentials and profiles
- [Usage Guide](USAGE.md) - Advanced features and workflows

## Developer Resources
- [Architecture](DEVELOPER.md) - Technical implementation details
- [Testing](TESTING.md) - Test coverage and approach
- [Deployment](DEPLOYMENT.md) - Building and publishing releases

## Key Features

### Command Line Usage
```bash
# Interactive mode
./cwlogs

# Specify profile and region
./cwlogs production us-west-2

# Use flags
./cwlogs --profile dev --region eu-west-1
```

### Navigation
- **Log Group Selection**: `↑↓/j/k` navigate, `Enter` select, `r` change region
- **Log Viewer**: `↑↓/j/k` scroll, `/` search, `b` back, `q` quit
- **Search**: `n/N` next/prev match, `Esc` clear

### Key Capabilities
- Real-time log streaming with follow mode
- Multi-region support with instant switching
- Fast regex search with highlighting
- Memory-bounded operation (5000 log buffer)
- Raw and formatted display modes
- Back navigation between log groups
