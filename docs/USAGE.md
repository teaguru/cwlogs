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
- **Type** - Start filtering log groups by name (automatic)
- `Enter` - Select log group
- `R` - Change AWS region
- `Esc` - Clear filter (if filtering) or quit
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
- `c` - Copy current log line to clipboard (original unformatted message)
- **Mouse selection** - Drag to select text, then Cmd+C/Ctrl+C to copy
- `b` - Back to log group selection
- `q` - Quit application

## Common Workflows

### Multi-Region Debugging
```bash
# Start with production profile
./cwlogs production

# In log group selection:
# 1. Just type "lambda" to filter only Lambda function logs
# 2. Press 'R' to change region if needed
# 4. Select us-west-2
# 5. Choose /aws/lambda/api-service
# 6. Debug issue in west coast
# 7. Press 'b' to go back
# 8. Press 'R' to switch to eu-west-1
# 9. Compare same service in Europe
```

### Log Group Filtering
```bash
# In log group selection screen:
# 1. Just start typing: "lambda" shows only Lambda logs
# 2. Continue typing "api" to further narrow to API-related logs
# 3. Press 'Backspace' to delete characters
# 4. Press 'Esc' to clear filter and show all groups
# 5. Use ↑↓ to navigate filtered results
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
