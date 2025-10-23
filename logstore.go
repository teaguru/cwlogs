package main

// logStore is a fixed-capacity ring buffer for log entries.
// Memory-bounded, O(1) append, no reallocation after init.
type logStore struct {
	entries  []logEntry
	start    int // index of oldest entry
	capacity int // max entries (fixed)
}

// newLogStore creates a ring buffer with fixed capacity
func newLogStore(capacity int) *logStore {
	return &logStore{
		entries:  make([]logEntry, 0, capacity),
		capacity: capacity,
	}
}

// Append adds a log entry. When full, overwrites oldest entry.
// Returns true if buffer wrapped (overwrote old entries).
func (s *logStore) Append(entry logEntry) bool {
	if len(s.entries) < s.capacity {
		// Still growing, just append
		s.entries = append(s.entries, entry)
		return false
	} else {
		// Full, overwrite oldest and advance start pointer
		s.entries[s.start] = entry
		s.start = (s.start + 1) % s.capacity
		return true // Buffer wrapped
	}
}

// Len returns the number of entries currently stored
func (s *logStore) Len() int {
	return len(s.entries)
}

// Slice returns a linearized copy of all entries in chronological order
func (s *logStore) Slice() []logEntry {
	if len(s.entries) < s.capacity {
		// Not full yet, return as-is
		return s.entries
	}

	// Full buffer, need to linearize circular structure
	result := make([]logEntry, s.capacity)
	copy(result, s.entries[s.start:])
	copy(result[s.capacity-s.start:], s.entries[:s.start])
	return result
}

// UpdateEntry updates an entry at the given index (for reprocessing)
func (s *logStore) UpdateEntry(index int, entry logEntry) {
	if index < 0 || index >= len(s.entries) {
		return // Bounds check
	}
	s.entries[index] = entry
}
