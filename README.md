# Distributed Fault-Tolerant Key-Value Store with Raft Consensus

This project implements a distributed, fault-tolerant key-value store using the Raft consensus algorithm in Go. The key-value store provides a reliable and consistent way to store and retrieve data across multiple nodes in a cluster, while the Raft consensus algorithm ensures data integrity and fault tolerance.

## Table of Contents

- [Introduction](#introduction)
- [Raft Consensus Algorithm](#raft-consensus-algorithm)
- [Key-Value Store](#key-value-store)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Configuration](#configuration)
- [Fault Tolerance](#fault-tolerance)
- [Performance](#performance)
- [Future Enhancements](#future-enhancements)
- [References](#references)
- [License](#license)

## Introduction

In distributed systems, ensuring data consistency and availability in the presence of node failures and network partitions is a challenging task. The Raft consensus algorithm, proposed by Diego Ongaro and John Ousterhout [1], provides a solution to this problem by maintaining a replicated log across multiple nodes and achieving consensus on the state of the system.

This project combines the Raft consensus algorithm with a key-value store to create a distributed, fault-tolerant data storage system. It allows clients to perform `Get`, `Set`, and `Delete` operations on key-value pairs, while the Raft algorithm ensures that the data remains consistent across all nodes in the cluster.

## Raft Consensus Algorithm

The Raft consensus algorithm is a leader-based algorithm that consists of three main components: leader election, log replication, and safety [1]. It operates on a set of nodes, where one node acts as the leader and the others are followers. The leader is responsible for managing the replication of the log entries to the followers and ensuring data consistency.

The Raft algorithm progresses through a series of terms, where each term has a unique identifier. During each term, a leader is elected through a randomized election process. Once a leader is elected, it starts accepting client requests and replicating the log entries to the followers. The followers apply the log entries to their local state machine in the same order as the leader.

In the event of a leader failure or network partition, the Raft algorithm automatically elects a new leader to maintain the availability of the system. The new leader ensures that all nodes have a consistent view of the log before proceeding with new client requests.

## Key-Value Store

The key-value store in this project provides a simple and efficient way to store and retrieve data. It supports the following operations:

- `Get(key)`: Retrieves the value associated with the specified key.
- `Set(key, value)`: Sets the value for the specified key.
- `Delete(key)`: Deletes the key-value pair associated with the specified key.

The key-value store is implemented using an in-memory map and supports concurrent access using locks. It also includes persistence capabilities, such as saving and loading snapshots of the store's state, to ensure data durability across node restarts.

## Architecture

The project is structured into several packages, each responsible for a specific functionality:

- `raft`: Implements the core functionality of the Raft consensus algorithm, including leader election, log replication, and state machine management.
- `kvstore`: Defines the key-value store interface and its implementation, providing methods for `Get`, `Set`, and `Delete` operations.
- `persistence`: Provides a persistent storage wrapper for the key-value store, handling snapshot saving and loading.
- `membership`: Manages the membership of nodes in the Raft cluster, keeping track of the current leader and cluster configuration.
- `snapshot`: Handles the storage and retrieval of snapshots based on the Raft log index.
- `server`: Implements the key-value store server, integrating the Raft node, key-value store, and snapshot mechanisms.
- `client`: Provides a client library for interacting with the key-value store server.

## Getting Started

To get started with the distributed key-value store, follow these steps:

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/distributed-kv-store.git
   ```

2. Install the dependencies:
   ```
   go mod download
   ```

3. Build the project:
   ```
   go build ./...
   ```

4. Start multiple instances of the key-value store server:
   ```
   ./kv-server --config config1.yaml
   ./kv-server --config config2.yaml
   ./kv-server --config config3.yaml
   ```

   Each server instance should have its own configuration file specifying the server's address, Raft configuration, and other relevant settings.

5. Use the client library to interact with the key-value store:
   ```go
   import "github.com/0xa48rx394r83e9/distributed-kv-store/client"

   func main() {
       c := client.NewKVClient("http://localhost:8080")

       err := c.Set("key1", "value1")
       if err != nil {
           log.Fatal(err)
       }

       value, err := c.Get("key1")
       if err != nil {
           log.Fatal(err)
       }
       fmt.Println(value)
   }
   ```

   The client library provides a simple interface to perform `Get`, `Set`, and `Delete` operations on the key-value store.

## Usage

The key-value store server exposes both RPC and HTTP interfaces for clients to interact with. The RPC interface is used for internal communication between Raft nodes, while the HTTP interface is intended for external clients.

To perform operations on the key-value store using the HTTP interface, you can use the following endpoints:

- `GET /get?key=<key>`: Retrieves the value associated with the specified key.
- `POST /set`: Sets the value for the specified key. The request body should contain a JSON object with "key" and "value" fields.
- `DELETE /delete?key=<key>`: Deletes the key-value pair associated with the specified key.

The client library provides a convenient way to interact with the key-value store using these HTTP endpoints.

## Configuration

The key-value store server can be configured using a YAML configuration file. The configuration file specifies various settings, such as the server's address, Raft configuration, and other relevant options.

Here's an example configuration file:

```yaml
rpcAddr: ":8081"
httpAddr: ":8080"

raft:
  nodeID: "node1"
  bindAddr: ":8082"
  advertisedAddr: "localhost:8082"
  dataDir: "data/node1"
  logLevel: "info"
  snapshotInterval: 30
  snapshotThreshold: 1024
  maxSnapshots: 5
```

The configuration file includes the following settings:

- `rpcAddr`: The address and port for the RPC server.
- `httpAddr`: The address and port for the HTTP server.
- `raft.nodeID`: The unique identifier for the Raft node.
- `raft.bindAddr`: The address and port for the Raft node to bind to.
- `raft.advertisedAddr`: The address and port advertised to other Raft nodes.
- `raft.dataDir`: The directory for storing Raft data and snapshots.
- `raft.logLevel`: The log level for the Raft node (e.g., "info", "debug", "warn").
- `raft.snapshotInterval`: The interval (in seconds) at which snapshots are taken.
- `raft.snapshotThreshold`: The threshold (in bytes) for triggering a snapshot.
- `raft.maxSnapshots`: The maximum number of snapshots to retain.

Make sure to provide a configuration file for each instance of the key-value store server.

## Fault Tolerance

The Raft consensus algorithm ensures fault tolerance by maintaining a replicated log across multiple nodes. In the event of a node failure or network partition, the algorithm automatically elects a new leader to maintain the availability of the system.

When a client sends a request to modify the key-value store (e.g., `Set` or `Delete` operation), the leader node receives the request, appends it to its log, and replicates the log entry to the follower nodes. Once a majority of nodes have acknowledged the log entry, it is considered committed, and the corresponding operation is applied to the key-value store.

The snapshots and log compaction mechanisms help in managing the size of the replicated log and enable efficient state synchronization between nodes. Snapshots capture the state of the key-value store at a particular point in time, allowing nodes to quickly catch up with the latest state without replaying the entire log.

## Performance

The performance of the distributed key-value store depends on various factors, such as the number of nodes in the cluster, network latency, and the size of the key-value pairs being stored.

The Raft consensus algorithm introduces some overhead due to the replication of log entries and the need for a majority of nodes to acknowledge each operation. However, it provides strong consistency guarantees and ensures data integrity in the presence of failures.

To optimize performance, the key-value store implements several techniques, such as batching of log entries, pipelining of client requests, and efficient snapshot management. These optimizations help in reducing the latency and improving the throughput of the system.

## Future Enhancements

There are several potential enhancements and optimizations that can be made to the distributed key-value store:

- Implement a more efficient storage engine, such as RocksDB or LevelDB, to improve the performance of the key-value store.
- Add support for multiple key spaces or namespaces to allow for logical separation of data.
- Implement a more advanced client library with features like connection pooling, load balancing, and failover.
- Explore the use of hardware acceleration, such as RDMA or NVMe, to further improve the performance of the system.
- Integrate with a service discovery mechanism, such as etcd or Consul, to simplify the process of adding and removing nodes from the cluster.
- Implement security features, such as encryption and access control, to ensure the confidentiality and integrity of the stored data.

## References

[1] Ongaro, D., & Ousterhout, J. (2014). In search of an understandable consensus algorithm. In 2014 USENIX Annual Technical Conference (USENIX ATC 14) (pp. 305-319).

[2] Howard, H., Schwarzkopf, M., Madhavapeddy, A., & Crowcroft, J. (2015). Raft refloated: Do we have consensus?. ACM SIGOPS Operating Systems Review, 49(1), 12-21.

## License

This project is licensed under the [MIT License](LICENSE).
