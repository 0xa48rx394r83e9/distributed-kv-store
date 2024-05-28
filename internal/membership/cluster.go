// membership/cluster.go
package membership

import (
    "sync"
)

type ClusterMembership struct {
    mu     sync.RWMutex
    nodes  map[string]bool
    leader string
}

func NewClusterMembership() *ClusterMembership {
    return &ClusterMembership{
        nodes: make(map[string]bool),
    }
}

func (c *ClusterMembership) AddNode(nodeID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.nodes[nodeID] = true
}

func (c *ClusterMembership) RemoveNode(nodeID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.nodes, nodeID)
}

func (c *ClusterMembership) IsNodePresent(nodeID string) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.nodes[nodeID]
}

func (c *ClusterMembership) GetNodes() []string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    nodes := make([]string, 0, len(c.nodes))
    for nodeID := range c.nodes {
        nodes = append(nodes, nodeID)
    }
    return nodes
}

func (c *ClusterMembership) SetLeader(nodeID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.leader = nodeID
}

func (c *ClusterMembership) GetLeader() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.leader
}

