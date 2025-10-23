# Usage Guide

## Command Line Options

### Basic Usage
```bash
# Interactive mode - select profile and log group
./cwlogs

# Specify profile
./cwlogs production

# Specify profile and region
./cwlogs production us-west-2

# Use flags (alternative syntax)
./cwlogs --profile dev --region eu-west-1
```

## Navigation and Controls

### Log Group Selection
- `↑↓` or `j/k` - Navigate through log groups
- `Enter` - Select log group
- `r` - Change AWS region
- `q` - Quit application

### Log Viewer
- `↑↓` or `j/k` - Scroll one line
- `Page Up/Down` - Fast scroll
- `g` - Go to top
- `G` or `End` - Go to bottom (enables follow mode)
- `/` - Start search
- `n/N` - Next/previous search match
- `Esc` - Clear search
- `J` - Toggle between Raw and Formatted modes
- `F` - Toggle follow mode (auto-scroll)
- `H` - Load more history
- `b` - Back to log group selection
- `q` - Quit application

## Common Workflows

### Multi-Region Debugging
```bash
# Start with production profile
./cwlogs production

# In log group selection:
# 1. Press 'r' to change region
# 2. Select us-west-2
# 3. Choose /aws/lambda/api-service
# 4. Debug issue in west coast
# 5. Press 'b' to go back
# 6. Press 'r' to switch to eu-west-1
# 7. Compare same service in Europe
```

### Development Monitoring
```bash
# Monitor development environment
./cwlogs dev

# Switch between services:
# 1. Select /aws/lambda/auth-service
# 2. Check authentication logs
# 3. Press 'b' to go back
# 4. Select /aws/lambda/api-service
# 5. Check API logs
```

### Incident Response
```bash
# Quick access to production logs
./cwlogs prod us-east-1

# Follow logs in real-time:
# 1. Select problematic service log group
# 2. Press 'F' to enable follow mode
# 3. Use '/' to search for error patterns
# 4. Press 'n/N' to navigate between matches
```

## Search Features

### Basic Search
- Press `/` to start search
- Type your search term (supports regex)
- Press `Enter` to execute
- Use `n` for next match, `N` for previous
- Press `Esc` to clear search

### Search Examples
```
# Find errors
/ERROR

# Find specific user
/user.*12345

# Find HTTP status codes
/[45][0-9][0-9]

# Find timestamps
/2024-.*14:30
```

## Display Modes

### Raw Mode
Shows original CloudWatch log messages:
```
[12:34:56] 2023/10/15 12:34:56 [INFO] User login successful
[12:34:57] 2023/10/15 12:34:57 [ERROR] Database connection failed
```

### Formatted Mode
Parses and colorizes structured logs:
```
[12:34:56] GET /api/users 200 2.3ms - Mozilla/5.0...
[12:34:57] POST /api/login 401 1.1ms - Invalid credentials
```

Toggle between modes with `J` key.

## Performance Tips

### Memory Management
- Application maintains only 5000 most recent logs
- Older logs automatically removed
- Memory usage stays constant (~50MB)

### Efficient Navigation
- Use follow mode (`F`) for real-time monitoring
- Search automatically disables follow mode
- Press `G` to re-enable follow and jump to latest logs

### Multi-Region Best Practices
- Change region in log group selection (not during viewing)
- Use consistent log group names across regions for easy comparison
- Bookmark frequently used profile/region combinations as shell aliases

## Troubleshooting

### Common Issues
- **No log groups found**: Check AWS credentials and region
- **Logs not updating**: Verify follow mode is enabled (`F`)
- **Search not working**: Clear search with `Esc` and try again
- **Performance issues**: Application handles thousands of logs efficiently

### AWS Configuration
- Ensure AWS credentials are configured: `aws configure`
- Check profile access: `aws sts get-caller-identity --profile myprofile`
- Verify CloudWatch permissions: `aws logs describe-log-groups --profile myprofile`
