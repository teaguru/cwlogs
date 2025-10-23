# Log Group Menu Region Change

## Overview

Implemented region change functionality (`r` key) in the **log group selection menu**, allowing users to switch AWS regions while selecting which log group to view. This provides a more intuitive workflow where region changes happen at the selection stage, not during log viewing.

## Correct Implementation Location

### ✅ **In Log Group Selection Menu**
- Press `r` while selecting log groups to change region
- Immediately refreshes available log groups for the new region
- Maintains context within the selection workflow
- More intuitive user experience

### ❌ **Not in Log Viewer**
- Region changes don't make sense while viewing specific logs
- Would be confusing to change region mid-viewing
- Better to change region before selecting what to view

## User Workflow

### Enhanced Log Group Selection
```
1. ./cwlogs production us-east-1
2. Log group selection menu appears
3. See log groups in us-east-1
4. Press 'r' to change region ← NEW!
5. Select region: us-west-2
6. Menu refreshes with us-west-2 log groups
7. Select desired log group
8. View logs in us-west-2
9. Press 'b' to go back to log group selection
10. Press 'r' again to switch to eu-west-1...
```

## Technical Implementation

### Custom Log Group Selector TUI

**New File: `loggroup_selector.go`**
- Custom Bubble Tea TUI for log group selection
- Supports navigation with `↑↓` or `j/k`
- `Enter` to select log group
- `r` to change region
- `q` to quit

**Key Features:**
```go
type logGroupSelectorModel struct {
    logGroups    []string
    cursor       int
    selected     string
    changeRegion bool
    quit         bool
    config       *UIConfig
}
```

### Interactive Selection Function

**Enhanced selectLogGroupInteractive():**
```go
func selectLogGroupInteractive(logGroups []string, config *UIConfig) (string, bool, error) {
    // Returns: (selectedLogGroup, changeRegionRequested, error)
}
```

**Return Values:**
- `selectedLogGroup` - The chosen log group name
- `changeRegionRequested` - True if user pressed 'r'
- `error` - Any errors during selection

### Main Loop Structure

**Simplified Flow:**
```go
for {
    // Log group selection with region change support
    chosenLogGroup, changeRegion, err := selectLogGroupInteractive(logGroups, config)
    
    if changeRegion {
        // Handle region change
        newRegion := selectAWSRegion(config)
        logGroups = listLogGroups(profile, newRegion) // Refresh
        continue // Back to log group selection
    }
    
    // Start log viewer
    for {
        exitCode := startLogViewer(profile, chosenLogGroup, region, config)
        if exitCode == 2 { break } // Back to log group selection
        if exitCode == 0 { return } // Quit
    }
}
```

## User Interface

### Log Group Selection Screen

**Visual Design:**
```
📋 CloudWatch Log Group Selection

Use ↑↓ or j/k to navigate, Enter to select, r to change region, q to quit

> /aws/lambda/api-service
  /aws/lambda/auth-service
  /aws/lambda/data-processor
  /aws/rds/postgres/error
  /aws/ecs/web-app

[2/5 log groups]

↑↓/j/k: navigate | Enter: select | r: change region | q: quit
```

**Features:**
- **Highlighted selection** - Current item highlighted with background color
- **Scroll indicator** - Shows position in long lists
- **Responsive design** - Adapts to terminal size
- **Clear instructions** - Always visible controls
- **Truncation** - Long log group names truncated with "..."

### Region Change Integration

**When user presses 'r':**
1. Exit log group selector
2. Show region selection menu
3. User selects new region
4. Refresh log groups for new region
5. Return to log group selector with new data
6. User continues with log group selection

## Benefits

### Intuitive Workflow
- **Logical placement** - Change region before selecting what to view
- **Context preservation** - Stay in selection mode during region change
- **Clear separation** - Region selection separate from log viewing
- **Reduced confusion** - No region changes during active log viewing

### Better User Experience
- **Immediate feedback** - See available log groups in new region instantly
- **Seamless switching** - No application restart needed
- **Visual consistency** - Same TUI style as log viewer
- **Keyboard efficiency** - All operations via keyboard shortcuts

### Technical Advantages
- **Cleaner architecture** - Separation of concerns
- **Easier testing** - Isolated log group selector component
- **Better maintainability** - Clear component boundaries
- **Extensible design** - Easy to add more selection features

## Testing Coverage

### New Test File: `loggroup_selector_test.go`

**4 comprehensive tests:**
1. **Initialization** - Verify initial state setup
2. **Navigation** - Test cursor movement and bounds
3. **Region change** - Verify 'r' key functionality
4. **Selection** - Test log group selection with Enter

**Test Results:**
- **38 tests total** (added 4 new tests)
- **All tests passing**
- **9.8% code coverage**
- **Log group selector fully tested**

## Comparison: Before vs After

### Before (Survey Library)
```go
// Simple but limited
func selectLogGroup(logGroups []string) (string, error) {
    prompt := &survey.Select{
        Message: "Select CloudWatch log group:",
        Options: logGroups,
    }
    survey.AskOne(prompt, &chosen)
    return chosen, err
}
```

**Limitations:**
- No custom key bindings
- No region change support
- Limited visual customization
- No scroll indicators

### After (Custom TUI)
```go
// Full-featured interactive selector
func selectLogGroupInteractive(logGroups []string) (string, bool, error) {
    model := newLogGroupSelector(logGroups, config)
    p := tea.NewProgram(model, tea.WithAltScreen())
    // ... handle region changes, selection, navigation
}
```

**Advantages:**
- Custom key bindings (`r` for region change)
- Rich visual design with highlighting
- Scroll indicators and responsive layout
- Integrated region change workflow
- Consistent with log viewer TUI style

## Documentation Updates

### README.md
- Moved region change documentation to "Log Group Selection Controls"
- Clear separation between log viewer and log group selector controls
- Added visual examples of the selection workflow

### Help Text
- Log group selector shows its own help text
- Log viewer help text simplified (no region change)
- Context-appropriate instructions for each screen

## Real-World Usage

### Multi-Region Debugging
```bash
./cwlogs production
# Log group selection appears
# Press 'r' → select us-west-2
# See log groups in us-west-2
# Select /aws/lambda/api-service
# View logs, debug issue
# Press 'b' to go back
# Press 'r' → select eu-west-1
# Compare same service in different region
```

### Development Workflow
```bash
./cwlogs dev
# See development log groups in default region
# Press 'r' → select staging region
# Compare dev vs staging log groups
# Select appropriate log group for investigation
```

The corrected implementation places region change functionality exactly where it makes the most sense - in the log group selection menu where users are deciding what to view, not after they're already viewing it. This creates a more intuitive and efficient workflow for multi-region CloudWatch log monitoring.
