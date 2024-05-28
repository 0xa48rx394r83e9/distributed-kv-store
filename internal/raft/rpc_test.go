// internal/raft/rpc_test.go
package raft

import (
    "testing"
)

func TestRequestVote(t *testing.T) {
    // Test case 1: Vote granted to candidate with higher term
    node := &RaftNode{
        currentTerm: 1,
        votedFor:    "",
        log:         NewRaftLog(),
    }
    args := &RequestVoteArgs{
        Term:         2,
        CandidateID:  "node2",
        LastLogIndex: 0,
        LastLogTerm:  0,
    }
    reply := &RequestVoteReply{}
    err := node.RequestVote(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !reply.VoteGranted {
        t.Errorf("expected vote to be granted")
    }
    if node.currentTerm != 2 {
        t.Errorf("expected node's term to be updated to 2")
    }
    if node.votedFor != "node2" {
        t.Errorf("expected node to vote for node2")
    }

    // Test case 2: Vote not granted to candidate with lower term
    node = &RaftNode{
        currentTerm: 3,
        votedFor:    "",
        log:         NewRaftLog(),
    }
    args = &RequestVoteArgs{
        Term:         2,
        CandidateID:  "node3",
        LastLogIndex: 0,
        LastLogTerm:  0,
    }
    reply = &RequestVoteReply{}
    err = node.RequestVote(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if reply.VoteGranted {
        t.Errorf("expected vote not to be granted")
    }
    if node.currentTerm != 3 {
        t.Errorf("expected node's term to remain unchanged")
    }
    if node.votedFor != "" {
        t.Errorf("expected node not to vote for any candidate")
    }

    // Test case 3: Vote not granted if already voted for another candidate
    node = &RaftNode{
        currentTerm: 2,
        votedFor:    "node2",
        log:         NewRaftLog(),
    }
    args = &RequestVoteArgs{
        Term:         2,
        CandidateID:  "node3",
        LastLogIndex: 0,
        LastLogTerm:  0,
    }
    reply = &RequestVoteReply{}
    err = node.RequestVote(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if reply.VoteGranted {
        t.Errorf("expected vote not to be granted")
    }
    if node.currentTerm != 2 {
        t.Errorf("expected node's term to remain unchanged")
    }
    if node.votedFor != "node2" {
        t.Errorf("expected node to keep its vote for node2")
    }
}

func TestAppendEntries(t *testing.T) {
    // Test case 1: Append entries to follower with matching previous log term and index
    node := &RaftNode{
        currentTerm: 2,
        log:         NewRaftLog(),
    }
    node.log.Append(LogEntry{Index: 1, Term: 1})
    node.log.Append(LogEntry{Index: 2, Term: 2})
    args := &AppendEntriesArgs{
        Term:         2,
        LeaderID:     "node1",
        PrevLogIndex: 1,
        PrevLogTerm:  1,
        Entries:      []LogEntry{{Index: 3, Term: 2}, {Index: 4, Term: 2}},
        LeaderCommit: 3,
    }
    reply := &AppendEntriesReply{}
    err := node.AppendEntries(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !reply.Success {
        t.Errorf("expected AppendEntries to succeed")
    }
    if node.log.LastIndex() != 4 {
        t.Errorf("expected last log index to be 4")
    }
    if node.commitIndex != 3 {
        t.Errorf("expected commit index to be 3")
    }

    // Test case 2: Reject AppendEntries with lower term
    node = &RaftNode{
        currentTerm: 3,
        log:         NewRaftLog(),
    }
    args = &AppendEntriesArgs{
        Term:         2,
        LeaderID:     "node1",
        PrevLogIndex: 0,
        PrevLogTerm:  0,
        Entries:      []LogEntry{{Index: 1, Term: 2}},
        LeaderCommit: 1,
    }
    reply = &AppendEntriesReply{}
    err = node.AppendEntries(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if reply.Success {
        t.Errorf("expected AppendEntries to be rejected")
    }
    if node.log.LastIndex() != 0 {
        t.Errorf("expected log to remain unchanged")
    }

    // Test case 3: Reject AppendEntries with mismatching previous log term or index
    node = &RaftNode{
        currentTerm: 2,
        log:         NewRaftLog(),
    }
    node.log.Append(LogEntry{Index: 1, Term: 1})
    args = &AppendEntriesArgs{
        Term:         2,
        LeaderID:     "node1",
        PrevLogIndex: 1,
        PrevLogTerm:  2,
        Entries:      []LogEntry{{Index: 2, Term: 2}},
        LeaderCommit: 1,
    }
    reply = &AppendEntriesReply{}
    err = node.AppendEntries(args, reply)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if reply.Success {
        t.Errorf("expected AppendEntries to be rejected")
    }
    if node.log.LastIndex() != 1 {
        t.Errorf("expected log to remain unchanged")
    }
}