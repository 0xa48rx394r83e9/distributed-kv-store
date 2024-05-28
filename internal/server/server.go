// server/server.go
package server

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "raft"
    "sync"
)

type KVServer struct {
    mu       sync.RWMutex
    raftNode *raft.RaftNode
    kvStore  raft.StateMachine
    rpcServer *rpc.Server
    httpServer *http.Server
}

func NewKVServer(raftNode *raft.RaftNode, kvStore raft.StateMachine) *KVServer {
    server := &KVServer{
        raftNode: raftNode,
        kvStore:  kvStore,
        rpcServer: rpc.NewServer(),
    }
    server.rpcServer.RegisterName("KVServer", server)
    return server
}

func (s *KVServer) Get(args *GetArgs, reply *GetReply) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.raftNode.State() != raft.Leader {
        return raft.ErrNotLeader
    }
    value, err := s.kvStore.Get(args.Key)
    if err != nil {
        return err
    }
    reply.Value = value
    return nil
}

func (s *KVServer) Set(args *SetArgs, reply *SetReply) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.raftNode.State() != raft.Leader {
        return raft.ErrNotLeader
    }
    cmd := &raft.Command{
        Operation: "set",
        Key:       args.Key,
        Value:     args.Value,
    }
    index, err := s.raftNode.Apply(cmd)
    if err != nil {
        return err
    }
    reply.Index = index
    return nil
}

func (s *KVServer) Delete(args *DeleteArgs, reply *DeleteReply) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.raftNode.State() != raft.Leader {
        return raft.ErrNotLeader
    }
    cmd := &raft.Command{
        Operation: "delete",
        Key:       args.Key,
    }
    index, err := s.raftNode.Apply(cmd)
    if err != nil {
        return err
    }
    reply.Index = index
    return nil
}

func (s *KVServer) StartRPC(listenAddr string) error {
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        return err
    }
    go s.rpcServer.Accept(listener)
    return nil
}

func (s *KVServer) StartHTTP(listenAddr string) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/get", s.handleGet)
    mux.HandleFunc("/set", s.handleSet)
    mux.HandleFunc("/delete", s.handleDelete)
    s.httpServer = &http.Server{
        Addr:    listenAddr,
        Handler: mux,
    }
    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil {
            log.Printf("HTTP server error: %v", err)
        }
    }()
    return nil
}

func (s *KVServer) handleGet(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    key := r.URL.Query().Get("key")
    if key == "" {
        http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
        return
    }

    args := &GetArgs{Key: key}
    var reply GetReply
    err := s.Get(args, &reply)
    if err != nil {
        if err == raft.ErrNotLeader {
            http.Error(w, err.Error(), http.StatusBadGateway)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(reply)
}

func (s *KVServer) handleSet(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var args SetArgs
    err := json.NewDecoder(r.Body).Decode(&args)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    var reply SetReply
    err = s.Set(&args, &reply)
    if err != nil {
        if err == raft.ErrNotLeader {
            http.Error(w, err.Error(), http.StatusBadGateway)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(reply)
}

func (s *KVServer) handleDelete(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    key := r.URL.Query().Get("key")
    if key == "" {
        http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
        return
    }

    args := &DeleteArgs{Key: key}
    var reply DeleteReply
    err := s.Delete(args, &reply)
    if err != nil {
        if err == raft.ErrNotLeader {
            http.Error(w, err.Error(), http.StatusBadGateway)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(reply)
}

type GetArgs struct {
    Key string
}

type GetReply struct {
    Value string
}

type SetArgs struct {
    Key   string
    Value string
}

type SetReply struct {
    Index int
}

type DeleteArgs struct {
    Key string
}

type DeleteReply struct {
    Index int
}