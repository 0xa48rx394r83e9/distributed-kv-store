// internal/raft/log.go
package raft

import (
    "sync"
)

type LogEntry struct {
    Index int
    Term  int
    Data  []byte
}

type RaftLog struct {
    mu      sync.RWMutex
    entries []LogEntry
}

func NewRaftLog() *RaftLog {
    return &RaftLog{
        entries: make([]LogEntry, 0),
    }
}

func (l *RaftLog) Append(entry LogEntry) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.entries = append(l.entries, entry)
}

func (l *RaftLog) GetEntry(index int) (LogEntry, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()
    if index < 0 || index >= len(l.entries) {
        return LogEntry{}, ErrIndexOutOfBounds
    }
    return l.entries[index], nil
}

func (l *RaftLog) LastIndex() int {
    l.mu.RLock()
    defer l.mu.RUnlock()
    if len(l.entries) == 0 {
        return 0
    }
    return l.entries[len(l.entries)-1].Index
}

func (l *RaftLog) Term(index int) (int, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()
    if index < 0 || index >= len(l.entries) {
        return 0, ErrIndexOutOfBounds
    }
    return l.entries[index].Term, nil
}

var ErrIndexOutOfBounds = errors.New("index out of bounds")

