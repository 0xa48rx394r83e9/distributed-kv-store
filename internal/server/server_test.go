// server/server_test.go
package server

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "raft"
    "testing"
)

type MockRaftNode struct {
    applyFunc func(cmd *raft.Command) (int, error)
    state     raft.State
}

func (m *MockRaftNode) Apply(cmd *raft.Command) (int, error) {
    return m.applyFunc(cmd)
}

func (m *MockRaftNode) State() raft.State {
    return m.state
}

type MockStateMachine struct {
    getFunc    func(key string) (string, error)
    setFunc    func(key, value string) error
    deleteFunc func(key string) error
}

func (m *MockStateMachine) Get(key string) (string, error) {
    return m.getFunc(key)
}

func (m *MockStateMachine) Set(key, value string) error {
    return m.setFunc(key, value)
}

func (m *MockStateMachine) Delete(key string) error {
    return m.deleteFunc(key)
}

func TestKVServer_Get(t *testing.T) {
    raftNode := &MockRaftNode{state: raft.Leader}
    stateMachine := &MockStateMachine{
        getFunc: func(key string) (string, error) {
            if key == "key1" {
                return "value1", nil
            }
            return "", KeyNotFoundError(key)
        },
    }
    server := NewKVServer(raftNode, stateMachine)

    req, _ := http.NewRequest("GET", "/get?key=key1", nil)
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(server.handleGet)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    var reply GetReply
    err := json.NewDecoder(rr.Body).Decode(&reply)
    if err != nil {
        t.Errorf("failed to decode response: %v", err)
    }
    if reply.Value != "value1" {
        t.Errorf("handler returned unexpected value: got %v want %v", reply.Value, "value1")
    }
}

func TestKVServer_Set(t *testing.T) {
    raftNode := &MockRaftNode{
        applyFunc: func(cmd *raft.Command) (int, error) {
            return 1, nil
        },
        state: raft.Leader,
    }
    stateMachine := &MockStateMachine{
        setFunc: func(key, value string) error {
            return nil
        },
    }
    server := NewKVServer(raftNode, stateMachine)

    args := SetArgs{Key: "key1", Value: "value1"}
    data, _ := json.Marshal(args)
    req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(data))
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(server.handleSet)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    var reply SetReply
    err := json.NewDecoder(rr.Body).Decode(&reply)
    if err != nil {
        t.Errorf("failed to decode response: %v", err)
    }
    if reply.Index != 1 {
        t.Errorf("handler returned unexpected index: got %v want %v", reply.Index, 1)
    }
}

func TestKVServer_Delete(t *testing.T) {
    raftNode := &MockRaftNode{
        applyFunc: func(cmd *raft.Command) (int, error) {
            return 1, nil
        },
        state: raft.Leader,
    }
    stateMachine := &MockStateMachine{
        deleteFunc: func(key string) error {
            return nil
        },
    }
    server := NewKVServer(raftNode, stateMachine)

    req, _ := http.NewRequest("DELETE", "/delete?key=key1", nil)
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(server.handleDelete)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    var reply DeleteReply
    err := json.NewDecoder(rr.Body).Decode(&reply)
    if err != nil {
        t.Errorf("failed to decode response: %v", err)
    }
    if reply.Index != 1 {
        t.Errorf("handler returned unexpected index: got %v want %v", reply.Index, 1)
    }
}

