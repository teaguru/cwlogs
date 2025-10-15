# Requirements Document

## Introduction

The CloudWatch Log Viewer needs three core fixes to be stable and reliable:

1. **Memory stays bounded** - No matter how long it runs
2. **Follow mode never breaks** - Cursor always tracks latest log
3. **Everything stays fast** - Search, toggle, rendering

That's it. Keep it simple.

## Core Requirements

### 1. Memory Stays Bounded

**What**: Keep max 5000 logs in memory, automatically drop oldest

**Why**: Long sessions shouldn't eat RAM

**Done when**:
- Memory usage flat after 1 hour of streaming
- Oldest logs automatically removed when buffer full
- No manual trimming logic needed

### 2. Follow Mode Never Breaks

**What**: When follow mode is on, cursor always points to latest log

**Why**: Users expect to see new logs immediately

**Done when**:
- New logs appear instantly when they arrive
- Cursor never points to invalid index
- Scrolling up disables follow, G/End re-enables it

### 3. Everything Stays Fast

**What**: No freezes, no lag, instant responses

**Why**: It's a CLI tool, should feel snappy

**Done when**:
- Format toggle (J): < 100ms
- Search: < 500ms on 5000 logs
- Rendering: smooth 60 FPS
- No blocking operations

## How We Know It Works

Run the viewer for 1 hour streaming logs:
- Memory usage stays flat
- Follow mode never loses track
- Format toggle feels instant
- Search works in both modes
- No crashes or freezes

That's the test.
