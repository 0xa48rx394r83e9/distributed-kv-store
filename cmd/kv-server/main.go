// cmd/kv-server/main.go
package main

import (
    "flag"
    "fmt"
    "kvstore"
    "log"
    "raft"
    "server"
)

func main() {
    configFile := flag.String("config", "config.yaml", "Path to the configuration file")
    flag.Parse()

    config, err := config.LoadConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    raftNode, err := raft.NewRaftNode(config.RaftConfig)
    if err != nil {
        log.Fatalf("Failed to create Raft node: %v", err)
    }

    kvStore := kvstore.NewKVStoreImpl()

    server := server.NewKVServer(raftNode, kvStore)

    err = server.StartRPC(config.RPCAddr)
    if err != nil {
        log.Fatalf("Failed to start RPC server: %v", err)
    }

    err = server.StartHTTP(config.HTTPAddr)
    if err != nil {
        log.Fatalf("Failed to start HTTP server: %v", err)
    }

    log.Printf("KV server started on %s (RPC) and %s (HTTP)", config.RPCAddr, config.HTTPAddr)

    select {}
}