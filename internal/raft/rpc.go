// raft/rpc.go
package raft

type RequestVoteArgs struct {
    Term         int
    CandidateID  string
    LastLogIndex int
    LastLogTerm  int
}

type RequestVoteReply struct {
    Term        int
    VoteGranted bool
}

type AppendEntriesArgs struct {
    Term         int
    LeaderID     string
    PrevLogIndex int
    PrevLogTerm  int
    Entries      []LogEntry
    LeaderCommit int
}

type AppendEntriesReply struct {
    Term    int
    Success bool
}