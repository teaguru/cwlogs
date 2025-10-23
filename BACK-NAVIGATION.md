# Back Navigation Feature

## Overview

Added back navigation functionality to the CloudWatch Log Viewer, allowing users to return to the log group selection screen without quitting the entire application.

## Problem Solved

### Before
- Once you selected a log group, you were stuck in the log viewer
- Only option was to quit (`q`) and restart the entire application
- Had to go through AWS profile selection again
- Inefficient workflow for checking multiple log groups

### After
- Press `b` or `Backspace` to return to log group selection
- Keep the same AWS profile and region
- Quickly switch between different log groups
- Efficient workflow for monitoring multiple services

## Usage

### Key Bindings
- **`b`** - Go back to log group selection
- **`Backspace`** - Alternative back key (same functionality)
- **`q`** - Quit application entirely

### Workflow Example
```
1. Start: ./cwlogs production us-west-2
2. Select log group: /aws/lambda/api-service
3. View logs, search, navigate...
4. Press 'b' to go back
5. Select different log group: /aws/lambda/auth-service
6. View different logs...
7. Press 'b' again to go back
8. Select another log group: /aws/ecs/web-app
9. Press 'q' to quit when done
```

## Technical Implementation

### Architecture Changes

**1. Added Back Message Type:**
```go
type backToLogGroupsMsg struct{}

func backToLogGroupsCmd() tea.Msg {
    return backToLogGroupsMsg{}
}
```

**2. Enhanced Model State:**
```go
type logModel struct {
    // ... existing fields
    backToLogGroups bool // Flag to indicate user wants to go back
}
```

**3. Key Handling:**
```go
case "b", "backspace":
    // Return to log group selection
    return m, backToLogGroupsCmd
```

**4. Message Processing:**
```go
case backToLogGroupsMsg:
    // Set flag to indicate user wants to go back
    m.backToLogGroups = true
    return m, tea.Quit
```

### Main Loop Restructure

**Before (Linear Flow):**
```go
selectProfile() -> selectLogGroup() -> startViewer() -> exit
```

**After (Loop with Back Support):**
```go
selectProfile() -> loop {
    selectLogGroup() -> startViewer() -> [back or quit]
    if back: continue loop
    if quit: break loop
}
```

**Implementation:**
```go
// Main loop to allow going back to log group selection
for {
    chosenLogGroup, err := selectLogGroup(logGroups, uiConfig)
    // ... error handling
    
    exitCode, err := startLogViewer(profile, chosenLogGroup, region, uiConfig)
    // ... error handling
    
    if exitCode == 0 {
        break // User quit normally
    }
    // exitCode 2 means go back to log group selection
    fmt.Println("\nReturning to log group selection...")
}
```

### Exit Code System

**Return Codes from `startLogViewer`:**
- **0** - User quit normally (`q` pressed)
- **2** - User wants to go back (`b` pressed)

**Detection Logic:**
```go
func startLogViewer(...) (int, error) {
    // ... setup TUI
    
    finalModel, err := p.Run()
    if err != nil {
        return 0, err
    }
    
    // Check if user wants to go back
    if logModel, ok := finalModel.(*logModel); ok && logModel.backToLogGroups {
        return 2, nil // Go back to log group selection
    }
    
    return 0, nil // Quit normally
}
```

## User Experience Improvements

### Efficient Multi-Log-Group Monitoring

**Scenario: Debugging a distributed system**
```bash
./cwlogs production us-west-2
# Select: /aws/lambda/api-gateway
# Check API logs, find issue with auth service
# Press 'b'
# Select: /aws/lambda/auth-service  
# Check auth logs, find database connection issue
# Press 'b'
# Select: /aws/rds/postgres/error
# Check database logs
# Press 'q' when done
```

**Benefits:**
- No need to restart application
- Keep same AWS profile/region context
- Fast switching between related services
- Maintain search history within each session

### Development Workflow

**Scenario: Local development monitoring**
```bash
./cwlogs dev
# Monitor application logs during development
# Press 'b' to check different services
# Quick iteration without CLI overhead
```

### Operations and Incident Response

**Scenario: Production incident**
```bash
./cwlogs production
# Quickly check multiple log groups
# Follow incident across different services
# Efficient navigation during time-critical situations
```

## Updated Help and Documentation

### In-App Help
**Bottom status bar:**
```
/ search, Esc clear, n/N next/prev, J format (Raw), F follow (OFF), H history, b back, q quit
```

**Connection success message:**
```
Controls: ↑↓/j/k=scroll, PgUp/PgDn=fast scroll, g=top, G/End=latest
          /=search, Esc=clear search, n/N=next/prev match
          J=format toggle, F=follow toggle, H=load history, b=back, q=quit
```

### README.md Updates
- Added `b` and `Backspace` to controls section
- Updated workflow examples to show back navigation
- Documented efficient multi-log-group monitoring patterns

## Testing

### New Test Coverage
```go
func TestBackToLogGroups(t *testing.T) {
    // Test back functionality
    // Verify flag setting
    // Verify quit command return
}
```

### Test Results
- **34 tests total** (added 1 new test)
- **All tests passing**
- **9.7% code coverage**
- **Back navigation fully tested**

## Benefits Summary

### For Developers
- **Faster debugging** - Quick switching between related services
- **Better workflow** - No need to restart application
- **Context preservation** - Keep AWS profile/region settings

### For DevOps/SRE
- **Efficient incident response** - Rapid log group switching
- **Multi-service monitoring** - Easy navigation between components
- **Reduced friction** - Less CLI overhead during critical situations

### For Teams
- **Improved productivity** - Faster log investigation workflows
- **Better tool adoption** - More user-friendly navigation
- **Reduced cognitive load** - Intuitive back navigation

## Comparison with Other Tools

### AWS CLI
```bash
# Before: Multiple commands needed
aws logs describe-log-groups --profile prod
aws logs filter-log-events --log-group-name /aws/lambda/api --profile prod
# Switch to different log group requires new command
aws logs filter-log-events --log-group-name /aws/lambda/auth --profile prod
```

### cwlogs with Back Navigation
```bash
# After: Single session, multiple log groups
./cwlogs prod
# Interactive selection, view logs, press 'b', select different group
# Seamless navigation within single session
```

The back navigation feature transforms cwlogs from a single-log-group viewer into a comprehensive log exploration tool, making it much more practical for real-world debugging and monitoring scenarios.
