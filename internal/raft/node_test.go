// raft/node_test.go
package raft

import (
    "testing"
    "time"
)

func TestNewRaftNode(t *testing.T) {
    config := &Config{
        NodeID:            "node1",
        Peers:             []string{"node2", "node3"},
        ElectionTimeout:   150 * time.Millisecond,
        HeartbeatInterval: 50 * time.Millisecond,
    }
    node, err := NewRaftNode(config)
    if err != nil {
        t.Fatalf("Failed to create Raft node: %v", err)
    }
    if node.id != "node1" {
        t.Errorf("Unexpected node ID. Expected 'node1', got '%s'", node.id)
    }
    if node.state != Follower {
        t.Errorf("Unexpected initial state. Expected Follower, got %v", node.state)
    }
}

func TestStartElection(t *testing.T) {
    config := &Config{
        NodeID:            "node1",
        Peers:             []string{"node2", "node3"},
        ElectionTimeout:   150 * time.Millisecond,
        HeartbeatInterval: 50 * time.Millisecond,
    }
    node, _ := NewRaftNode(config)
    node.startElection()
    if node.state != Candidate {
        t.Errorf("Node should have transitioned to Candidate state")
    }
    if node.currentTerm != 1 {
        t.Errorf("Current term should have incremented. Expected 1, got %d", node.currentTerm)
    }
}

func TestBecomeFollower(t *testing.T) {
    config := &Config{
        NodeID:            "node1",
        Peers:             []string{"node2", "node3"},
        ElectionTimeout:   150 * time.Millisecond,
        HeartbeatInterval: 50 * time.Millisecond,
    }
    node, _ := NewRaftNode(config)
    node.becomeFollower(2)
    if node.state != Follower {
        t.Errorf("Node should have transitioned to Follower state")
    }
    if node.currentTerm != 2 {
        t.Errorf("Current term should have updated to 2, got %d", node.currentTerm)
    }
    if node.votedFor != "" {
        t.Errorf("Voted for should have been reset, got %s", node.votedFor)
    }
}

func TestBecomeLeader(t *testing.T) {
    config := &Config{
        NodeID:            "node1",
        Peers:             []string{"node2", "node3"},
        ElectionTimeout:   150 * time.Millisecond,
        HeartbeatInterval: 50 * time.Millisecond,
    }
    node, _ := NewRaftNode(config)
    node.state = Candidate
    node.becomeLeader()
    if node.state != Leader {
        t.Errorf("Node should have transitioned to Leader state")
    }
    if len(node.nextIndex) != 2 {
        t.Errorf("Next index map should have been initialized for all peers")
    }
    if len(node.matchIndex) != 2 {
        t.Errorf("Match index map should have been initialized for all peers")
    }
}

