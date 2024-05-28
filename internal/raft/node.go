// raft/node.go
package raft

import (
    "fmt"
    "log"
    "math/rand"
    "net/rpc"
    "sync"
    "time"
)

type RaftState int

const (
    Follower RaftState = iota
    Candidate
    Leader
)

type RaftNode struct {
    mu sync.RWMutex
    id string
    state RaftState
    currentTerm int
    votedFor    string
    logEntries  []LogEntry
    commitIndex int
    lastApplied int
    nextIndex   map[string]int
    matchIndex  map[string]int
    electionTimer *time.Timer
    heartbeatTimer *time.Timer
    config *Config
}

func NewRaftNode(config *Config) (*RaftNode, error) {
    node := &RaftNode{
        id: config.NodeID,
        state: Follower,
        currentTerm: 0,
        votedFor: "",
        logEntries: make([]LogEntry, 0),
        commitIndex: 0,
        lastApplied: 0,
        nextIndex: make(map[string]int),
        matchIndex: make(map[string]int),
        config: config,
    }
    node.resetElectionTimer()
    return node, nil
}

func (n *RaftNode) resetElectionTimer() {
    n.electionTimer = time.AfterFunc(randTimeDuration(n.config.ElectionTimeout), func() {
        n.startElection()
    })
}

func (n *RaftNode) startElection() {
    n.mu.Lock()
    defer n.mu.Unlock()

    if n.state == Leader {
        return
    }

    n.state = Candidate
    n.currentTerm++
    n.votedFor = n.id
    n.resetElectionTimer()

    lastLogIndex := len(n.logEntries) - 1
    lastLogTerm := 0
    if lastLogIndex >= 0 {
        lastLogTerm = n.logEntries[lastLogIndex].Term
    }

    voteCount := 1
    for _, peer := range n.config.Peers {
        go func(peer string) {
            args := &RequestVoteArgs{
                Term:         n.currentTerm,
                CandidateID:  n.id,
                LastLogIndex: lastLogIndex,
                LastLogTerm:  lastLogTerm,
            }
            var reply RequestVoteReply
            if err := n.sendRequestVote(peer, args, &reply); err != nil {
                log.Printf("Failed to send RequestVote to %s: %v", peer, err)
                return
            }
            n.mu.Lock()
            defer n.mu.Unlock()
            if reply.Term > n.currentTerm {
                n.becomeFollower(reply.Term)
            } else if reply.VoteGranted {
                voteCount++
                if voteCount > len(n.config.Peers)/2 {
                    n.becomeLeader()
                }
            }
        }(peer)
    }
}

func (n *RaftNode) becomeFollower(term int) {
    n.mu.Lock()
    defer n.mu.Unlock()

    if term > n.currentTerm {
        n.currentTerm = term
        n.state = Follower
        n.votedFor = ""
        n.resetElectionTimer()
    }
}

func (n *RaftNode) becomeLeader() {
    n.mu.Lock()
    defer n.mu.Unlock()

    if n.state != Candidate {
        return
    }

    n.state = Leader
    n.nextIndex = make(map[string]int)
    n.matchIndex = make(map[string]int)
    for _, peer := range n.config.Peers {
        n.nextIndex[peer] = len(n.logEntries)
        n.matchIndex[peer] = 0
    }
    n.startHeartbeat()
}

func (n *RaftNode) startHeartbeat() {
    n.heartbeatTimer = time.AfterFunc(n.config.HeartbeatInterval, func() {
        n.sendHeartbeat()
        n.startHeartbeat()
    })
}

func (n *RaftNode) sendHeartbeat() {
    n.mu.Lock()
    defer n.mu.Unlock()

    for _, peer := range n.config.Peers {
        go func(peer string) {
            prevLogIndex := n.nextIndex[peer] - 1
            prevLogTerm := 0
            if prevLogIndex >= 0 {
                prevLogTerm = n.logEntries[prevLogIndex].Term
            }
            entries := n.logEntries[n.nextIndex[peer]:]
            args := &AppendEntriesArgs{
                Term:         n.currentTerm,
                LeaderID:     n.id,
                PrevLogIndex: prevLogIndex,
                PrevLogTerm:  prevLogTerm,
                Entries:      entries,
                LeaderCommit: n.commitIndex,
            }
            var reply AppendEntriesReply
            if err := n.sendAppendEntries(peer, args, &reply); err != nil {
                log.Printf("Failed to send AppendEntries to %s: %v", peer, err)
                return
            }
            n.mu.Lock()
            defer n.mu.Unlock()
            if reply.Term > n.currentTerm {
                n.becomeFollower(reply.Term)
            } else if reply.Success {
                n.nextIndex[peer] = len(n.logEntries)
                n.matchIndex[peer] = n.nextIndex[peer] - 1
            } else {
                n.nextIndex[peer]--
            }
        }(peer)
    }
}

func (n *RaftNode) sendRequestVote(peer string, args *RequestVoteArgs, reply *RequestVoteReply) error {
    client, err := rpc.Dial("tcp", peer)
    if err != nil {
        return fmt.Errorf("failed to dial peer %s: %v", peer, err)
    }
    defer client.Close()

    err = client.Call("RaftNode.RequestVote", args, reply)
    if err != nil {
        return fmt.Errorf("failed to send RequestVote RPC to peer %s: %v", peer, err)
    }

    return nil
}

func (n *RaftNode) sendAppendEntries(peer string, args *AppendEntriesArgs, reply *AppendEntriesReply) error {
    client, err := rpc.Dial("tcp", peer)
    if err != nil {
        return fmt.Errorf("failed to dial peer %s: %v", peer, err)
    }
    defer client.Close()

    err = client.Call("RaftNode.AppendEntries", args, reply)
    if err != nil {
        return fmt.Errorf("failed to send AppendEntries RPC to peer %s: %v", peer, err)
    }

    return nil
}

func randTimeDuration(base time.Duration) time.Duration {
    return base + time.Duration(rand.Int63n(int64(base)))
}

// RaftNode RPC methods

func (n *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
    n.mu.Lock()
    defer n.mu.Unlock()

    if args.Term < n.currentTerm {
        reply.Term = n.currentTerm
        reply.VoteGranted = false
        return nil
    }

    if args.Term > n.currentTerm {
        n.becomeFollower(args.Term)
    }

    if n.votedFor == "" || n.votedFor == args.CandidateID {
        lastLogIndex := len(n.logEntries) - 1
        lastLogTerm := 0
        if lastLogIndex >= 0 {
            lastLogTerm = n.logEntries[lastLogIndex].Term
        }
        if args.LastLogTerm > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex) {
            n.votedFor = args.CandidateID
            reply.VoteGranted = true
        }
    }

    reply.Term = n.currentTerm
    return nil
}

func (n *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
    n.mu.Lock()
    defer n.mu.Unlock()

    if args.Term < n.currentTerm {
        reply.Term = n.currentTerm
        reply.Success = false
        return nil
    }

    if args.Term > n.currentTerm {
        n.becomeFollower(args.Term)
    }

    n.resetElectionTimer()

    if args.PrevLogIndex >= len(n.logEntries) {
        reply.Term = n.currentTerm
        reply.Success = false
        return nil
    }

    if args.PrevLogIndex >= 0 && n.logEntries[args.PrevLogIndex].Term != args.PrevLogTerm {
        reply.Term = n.currentTerm
        reply.Success = false
        return nil
    }

    // Append new entries
    n.logEntries = n.logEntries[:args.PrevLogIndex+1]
    n.logEntries = append(n.logEntries, args.Entries...)

    // Update commit index
    if args.LeaderCommit > n.commitIndex {
        n.commitIndex = min(args.LeaderCommit, len(n.logEntries)-1)
    }

    reply.Term = n.currentTerm
    reply.Success = true
    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

