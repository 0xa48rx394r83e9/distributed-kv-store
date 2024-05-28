// client/client_test.go
package client

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestKVClient_Get(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            t.Errorf("expected GET request, got %s", r.Method)
        }
        if r.URL.Path != "/get" {
            t.Errorf("expected /get, got %s", r.URL.Path)
        }
        key := r.URL.Query().Get("key")
        if key != "key1" {
            t.Errorf("expected key1, got %s", key)
        }
        reply := GetReply{Value: "value1"}
        json.NewEncoder(w).Encode(reply)
    }))
    defer server.Close()

    client := NewKVClient(server.URL)
    value, err := client.Get("key1")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if value != "value1" {
        t.Errorf("expected value1, got %s", value)
    }
}

func TestKVClient_Set(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            t.Errorf("expected POST request, got %s", r.Method)
        }
        if r.URL.Path != "/set" {
            t.Errorf("expected /set, got %s", r.URL.Path)
        }
        var args SetArgs
        err := json.NewDecoder(r.Body).Decode(&args)
        if err != nil {
            t.Errorf("failed to decode request body: %v", err)
        }
        if args.Key != "key1" || args.Value != "value1" {
            t.Errorf("unexpected key-value pair: got %v want {\"key\":\"key1\",\"value\":\"value1\"}", args)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    client := NewKVClient(server.URL)
    err := client.Set("key1", "value1")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestKVClient_Delete(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "DELETE" {
            t.Errorf("expected DELETE request, got %s", r.Method)
        }
        if r.URL.Path != "/delete" {
            t.Errorf("expected /delete, got %s", r.URL.Path)
        }
        key := r.URL.Query().Get("key")
        if key != "key1" {
            t.Errorf("expected key1, got %s", key)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    client := NewKVClient(server.URL)
    err := client.Delete("key1")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}