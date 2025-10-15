# Implementation Plan

## Goal

Make the log viewer stable and fast with minimal code changes.

## Tasks

- [x] 1. Implement and integrate ring buffer
  - Create `logstore.go` with fixed-size circular buffer (~40 lines)
  - Implement `NewLogStore(capacity)`, `Append()`, `Len()`, `Slice()`
  - Use fixed capacity to prevent slice growth
  - Add `store *LogStore` field to logModel
  - Initialize in `startLogViewer()` with capacity 5000
  - Add `safeLogs()` helper that never panics
  - Replace `m.logs = append(...)` with `m.store.Append(...)`
  - Replace `m.logs` reads with `m.safeLogs()`
  - Remove old trimming logic
  - _Requirements: 1, 2_

- [x] 2. Fix cursor management
  - Add `fixCursor()` method using `safeLogs()`
  - Call after every log append
  - Call after cursor movement (up/down/page/search)
  - Handle empty buffer gracefully (cursor = 0)
  - Handle follow mode: if enabled, cursor = last index
  - Handle bounds: clamp cursor to [0, len-1]
  - _Requirements: 2_

- [x] 3. Add View() safety checks
  - Add optional `defer recover()` at start for crash-proof rendering
  - Check width/height > 0, return "Initializing..." if invalid
  - Use `safeLogs()` to get logs safely
  - Check for empty buffer, return "No logs yet" if empty
  - Optional: pre-grow strings.Builder to 4KB for smoother rendering
  - _Requirements: 3_

- [x] 4. Optimize format toggle
  - Ensure toggleFormat() returns tea.Cmd for redraw
  - Rebuild all entries from OriginalMessage
  - Clear search state on toggle
  - Don't call fixCursor() (toggle doesn't change log count)
  - _Requirements: 3_

- [ ] 5. Test stability
  - Run for 30 minutes, monitor memory in Activity Monitor
  - Verify follow mode stays correct when new logs arrive
  - Toggle format repeatedly (J key)
  - Search in both raw and formatted modes
  - Simulate > 5000 logs to verify oldest entries roll over
  - _Requirements: 1, 2, 3_

## That's It

5 tasks. Single session. Keep it simple.
