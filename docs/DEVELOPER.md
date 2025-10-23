# CloudWatch Log Viewer

A high-performance, real-time CloudWatch log viewer built with Go and Bubble Tea TUI framework.

## Overview

This application provides a terminal-based interface for viewing AWS CloudWatch logs with features like real-time streaming, search, format toggling, and memory-bounded operation for long-running sessions.

## Key Features

- **Real-time log streaming** with configurable refresh intervals
- **Memory-bounded operation** using a ring buffer (5000 log capacity)
- **Dual display modes**: Raw CloudWatch output and formatted/colorized logs
- **Fast search** with regex support and highlight caching
- **Follow mode** for automatic scrolling to new logs
- **Word wrapping** for long log lines
- **Stable cursor management** with viewport centering

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                         LogModel                             │
│  (Main TUI state, orchestrates all components)              │
└────────────┬────────────────────────────────────────────────┘
             │
    ┌────────┼────────┬────────────┬──────────────┐
    │        │        │            │              │
    ▼        ▼        ▼            ▼              ▼
┌────────┐ ┌──────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│LogStore│ │Parser│ │Search    │ │Cursor    │ │Renderer  │
│        │ │      │ │Cache     │ │Manager   │ │          │
└────────┘ └──────┘ └──────────┘ └──────────┘ └──────────┘
```

### File Structure

```
├── main.go              # Entry point, AWS profile selection, initialization
├── model.go             # Core TUI model, Update() method, message handling
├── model_methods.go     # View() rendering, search logic, highlight caching
├── logstore.go          # Ring buffer implementation for memory management
├── parser.go            # Log parsing, formatting (raw/formatted modes)
├── config.go            # Configuration, styling, UI settings
├── ui.go                # User interface helpers, welcome messages
└── aws.go               # AWS CloudWatch API integration
```

## Core Architecture Patterns

### 1. Ring Buffer (LogStore)

**Purpose**: Memory-bounded log storage with O(1) append and automatic trimming.

```go
type LogStore struct {
    entries  []LogEntry
    start    int // Index of oldest entry
    capacity int // Fixed capacity (5000)
}
```

**Key Features**:
- Fixed memory footprint regardless of session duration
- Circular indexing with modulo arithmetic
- Automatic overwrite of oldest entries when full
- Returns `true` from `Append()` when buffer wraps (for search invalidation)

**Why Ring Buffer**:
- Prevents memory leaks in long-running sessions
- Maintains constant performance (O(1) operations)
- Provides predictable memory usage

### 2. Bubble Tea Architecture

**Message-Driven Updates**:
```go
type logModel struct {
    // State
    store     *LogStore
    cursor    int
    followMode bool
    searchQuery string
    // ... other fields
}

func (m *logModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *logModel) View() string
```

**Key Messages**:
- `tickMsg`: Periodic log fetching (every 5 seconds)
- `logsWithTokenMsg`: New logs from CloudWatch API
- `delayedSearchMsg`: Delayed search after buffer rollover
- `loadingMsg`: Loading state updates

### 3. Search and Highlight System

**Two-Phase Approach**:
1. **Search Phase**: Find matches and store indices
2. **Highlight Phase**: Precompute highlighted strings for matched lines

```go
// Search finds matches and stores indices
func (m *logModel) performSearch() {
    logs := m.safeLogs()
    for i, log := range logs {
        if m.searchRegex.MatchString(stripANSI(log.Raw)) {
            m.matches = append(m.matches, i)
        }
    }
    m.applyHighlights() // Precompute highlights
}

// Highlights are cached for performance
m.highlighted[idx] = highlightedString
```

**ANSI-Safe Search**:
- Uses `stripANSI()` to remove color codes before regex matching
- Preserves original formatting in highlight cache
- Prevents cursor styling from overwriting search highlights

### 4. Cursor and Viewport Management

**Follow Mode Logic**:
```go
func (m *logModel) fixCursor() {
    logs := m.safeLogs()
    // Clamp to valid range
    if m.cursor >= len(logs) { m.cursor = len(logs) - 1 }
    if m.cursor < 0 { m.cursor = 0 }
    
    // Follow mode: always track latest
    if m.followMode { m.cursor = len(logs) - 1 }
}
```

**Viewport Centering**:
```go
// View() calculates visible range around cursor
start := m.cursor - (m.height-6)/2
end := start + (m.height - 6)
```

### 5. Format Toggle System

**Dual Mode Support**:
- **Raw Mode**: `strings.TrimSpace()` only (fast path)
- **Formatted Mode**: Access log parsing + JSON pretty-printing + colorization

```go
func formatLogMessage(message string, config *UIConfig) string {
    // Fast path for raw mode
    if !config.ParseAccessLogs && !config.PrettyPrintJSON {
        return strings.TrimSpace(message)
    }
    // Full processing for formatted mode
    // ... parsing and formatting logic
}
```

## Performance Optimizations

### 1. Memory Management
- **Ring Buffer**: Fixed 5000-entry capacity prevents unbounded growth
- **Lazy Formatting**: Logs formatted only when visible or when mode changes
- **Bounded Caching**: Search highlights limited to matched entries only

### 2. Rendering Optimizations
- **Viewport Culling**: Only render visible log lines
- **Pre-grown Buffers**: `strings.Builder` with 4KB initial capacity
- **ANSI-Safe Operations**: Efficient color code handling without breaking formatting

### 3. Search Performance
- **Highlight Caching**: Precompute highlights once, reuse on navigation
- **Incremental Updates**: Only refresh current match highlight on n/N navigation
- **Query Deduplication**: Skip reprocessing identical search queries

### 4. Real-time Streaming
- **Adaptive Fetch Windows**: 1-2 minute windows for near real-time updates
- **Follow Mode Integration**: Automatic cursor positioning for new logs
- **Fetch Coalescing**: Prevent concurrent CloudWatch requests

## Critical Edge Cases Handled

### 1. Ring Buffer Rollover
**Problem**: When buffer wraps, search indices become invalid.

**Solution**: 
```go
// Detect wrap in Append()
func (s *LogStore) Append(entry LogEntry) bool {
    if len(s.entries) < s.capacity {
        s.entries = append(s.entries, entry)
        return false // No wrap
    }
    // Buffer full - overwrite and signal wrap
    s.entries[s.start] = entry
    s.start = (s.start + 1) % s.capacity
    return true // Wrapped
}

// Handle wrap with delayed re-search
if wrapped {
    m.clearSearchState()
    // Schedule delayed search for next frame
    return m, delayedSearchMsg{oldQuery}
}
```

### 2. Search Highlight Conflicts
**Problem**: Cursor styling overwrites search highlights.

**Solution**:
```go
if i == m.cursor && j == 0 {
    if _, hasHighlight := m.highlighted[i]; hasHighlight {
        rendered = "▌ " + sub // Indicator only, preserve highlights
    } else {
        rendered = m.config.CursorStyle().Render(sub) // Full cursor style
    }
}
```

### 3. Format Toggle Consistency
**Problem**: Search highlights become stale when switching raw/formatted modes.

**Solution**:
```go
func (m *logModel) toggleFormat() tea.Cmd {
    // ... toggle logic
    
    // Recalculate highlights for new format
    if len(m.matches) > 0 {
        m.applyHighlights()
    }
    
    return redrawCmd
}
```

### 4. Follow Mode Race Conditions
**Problem**: Background ticks override search cursor positioning.

**Solution**:
```go
// Disable follow mode during search navigation
func (m *logModel) centerOnCursor() {
    m.followMode = false // Stop following
    m.fixCursor()        // Ensure valid position
}
```

## Configuration

### Key Settings
```go
type UIConfig struct {
    RefreshInterval  int  // Auto-refresh interval (seconds)
    LogsPerFetch    int  // Logs per CloudWatch API call
    MaxLogBuffer    int  // Ring buffer capacity
    LogTimeRange    int  // Initial time range (hours)
    ParseAccessLogs bool // Enable formatted mode
    PrettyPrintJSON bool // Enable JSON formatting
}
```

### Performance Tuning
- **Buffer Capacity**: Default 5000 entries (~50MB for typical logs)
- **Fetch Window**: 1-2 minutes for real-time, 2-10 minutes for history
- **Refresh Rate**: 5 seconds (configurable)

## Development Guidelines

### Adding New Features
1. **Follow Message-Driven Pattern**: Use Bubble Tea messages for state changes
2. **Maintain Ring Buffer Compatibility**: Consider wrap detection for stateful features
3. **Preserve Performance**: Use caching and lazy evaluation where possible
4. **Handle Edge Cases**: Test with buffer rollover, format toggling, and search

### Testing Scenarios
1. **Long Sessions**: Run for 1+ hours, verify memory stability
2. **Buffer Rollover**: Generate >5000 logs, test search functionality
3. **Format Toggle**: Switch modes during active search
4. **Follow Mode**: Verify cursor tracking with rapid log generation

### Common Pitfalls
- **Direct `m.logs` Access**: Always use `m.safeLogs()` for current buffer state
- **Search Index Staleness**: Clear search state on buffer modifications
- **Cursor Bounds**: Always validate cursor position after log changes
- **ANSI Code Handling**: Use `stripANSI()` for text operations, preserve for display

## Dependencies

- **Bubble Tea**: TUI framework for terminal interfaces
- **Lipgloss**: Styling and layout for terminal output
- **AWS SDK v2**: CloudWatch Logs API integration

## Performance Characteristics

- **Memory Usage**: O(buffer_capacity) - typically ~50MB
- **Search Time**: O(buffer_size × query_complexity) - ~200ms for 5000 logs
- **Render Time**: O(viewport_height) - <16ms for 60 FPS
- **Append Time**: O(1) - constant time regardless of buffer size

This architecture provides a robust, performant foundation for real-time log viewing with predictable resource usage and responsive user interaction.
