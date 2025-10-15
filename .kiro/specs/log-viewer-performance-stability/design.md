# Design Document

## The Problem

Current code has three issues:
1. `m.logs` grows forever → memory leak
2. Cursor gets out of sync when we trim logs → follow mode breaks
3. Format toggle reprocesses everything → slow

## The Solution

### 1. Ring Buffer for Log Storage

Replace `logs []LogEntry` with a simple ring buffer:

```go
type LogStore struct {
    entries  []LogEntry
    start    int  // oldest entry index
    capacity int  // max (5000)
}

func NewLogStore(capacity int) *LogStore {
    return &LogStore{
        entries:  make([]LogEntry, 0, capacity),
        capacity: capacity,
    }
}

func (s *LogStore) Append(entry LogEntry) {
    if len(s.entries) < s.capacity {
        s.entries = append(s.entries, entry)
    } else {
        s.entries[s.start] = entry
        s.start = (s.start + 1) % s.capacity
    }
}

func (s *LogStore) Len() int {
    return len(s.entries)
}

func (s *LogStore) Slice() []LogEntry {
    if len(s.entries) < s.capacity {
        return s.entries
    }
    // Linearize circular buffer
    result := make([]LogEntry, s.capacity)
    copy(result, s.entries[s.start:])
    copy(result[s.capacity-s.start:], s.entries[:s.start])
    return result
}
```

**Why it's bulletproof**:
- Fixed capacity prevents slice growth
- No reallocation after initial buffer
- Deterministic memory footprint
- O(1) append, always

### 2. Fix Cursor Management

Add two simple helpers:

```go
// safeLogs returns logs safely, never panics
func (m *logModel) safeLogs() []LogEntry {
    if m.store == nil {
        return nil
    }
    return m.store.Slice()
}

// fixCursor keeps cursor in valid range
func (m *logModel) fixCursor() {
    logs := m.safeLogs()
    if len(logs) == 0 {
        m.cursor = 0
        return
    }
    if m.cursor >= len(logs) {
        m.cursor = len(logs) - 1
    }
    if m.cursor < 0 {
        m.cursor = 0
    }
    if m.followMode {
        m.cursor = len(logs) - 1
    }
}
```

Call `fixCursor()` after every log append or cursor movement. Done.

### 3. Smart Format Toggle

Rebuild from `OriginalMessage` (fast because raw mode just trims whitespace):

```go
func (m *logModel) toggleFormat() tea.Cmd {
    m.config.ParseAccessLogs = !m.config.ParseAccessLogs
    m.config.PrettyPrintJSON = m.config.ParseAccessLogs
    
    // Rebuild all entries from original messages
    for i := range m.logs {
        entry := makeLogEntry(
            m.logs[i].Timestamp,
            m.logs[i].OriginalMessage,
            m.config,
        )
        m.logs[i].Message = entry.Message
        m.logs[i].Raw = entry.Raw
    }
    
    // Clear search since text changed
    m.clearSearchState()
    
    return func() tea.Msg { return nil }  // Trigger redraw
}
```

No need to call `fixCursor()` - toggle doesn't change log count.

## What We're NOT Doing

- ❌ No separate CursorManager struct
- ❌ No SearchIndex component
- ❌ No Formatter interface
- ❌ No feature flags
- ❌ No telemetry system
- ❌ No event deduplication (CloudWatch doesn't send dupes in practice)
- ❌ No complex caching strategies

## What We ARE Doing

- ✅ Ring buffer (one file: `logstore.go`, ~40 lines)
- ✅ Fix cursor after every change
- ✅ Keep search ANSI-safe (already works)
- ✅ Make toggle fast (already mostly works)
- ✅ Add bounds checks in View()
- ✅ Add `safeLogs()` helper to prevent panics
- ✅ Optional: panic guard in View() for crash-proof rendering

## Files Changed

```
logstore.go          (new) - Ring buffer (~40 lines)
model.go             - Add LogStore, safeLogs() helper
model_methods.go     - Call fixCursor() after appends, add panic guard
parser.go            - No changes needed
```

That's it. ~150 lines of new code total.

## AK-47 Grade Reliability

**Combat-tested patterns**:
1. Fixed-size ring buffer - no slice growth, no reallocation
2. `safeLogs()` helper - never panics, even during init
3. `fixCursor()` - always valid, no out-of-bounds
4. Optional `defer recover()` in View() - skip frame instead of crash
5. Pre-grown string builder - smoother rendering, less GC

**Result**: Stable under load, predictable, tiny, easy to debug.

## Migration

1. Create `logstore.go` with ring buffer
2. Add `store *LogStore` to model
3. Replace `m.logs = append(...)` with `m.store.Append(...)`
4. Replace `m.logs` access with `m.store.Slice()`
5. Call `m.fixCursor()` after appends
6. Done

## Testing

Manual test:
1. Run viewer for 30 minutes
2. Watch memory in Activity Monitor
3. Verify follow mode tracks new logs
4. Toggle format a bunch of times
5. Search in both modes

If it works, ship it.
