// kvstore/store_test.go
package kvstore

import (
    "testing"
)

func TestKVStoreImpl_Get(t *testing.T) {
    store := NewKVStoreImpl()
    store.Set("key1", "value1")

    value, err := store.Get("key1")
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if value != "value1" {
        t.Errorf("Expected value 'value1', got '%s'", value)
    }

    _, err = store.Get("nonexistentKey")
    if err == nil {
        t.Error("Expected KeyNotFoundError, got nil")
    }
}

func TestKVStoreImpl_Set(t *testing.T) {
    store := NewKVStoreImpl()

    err := store.Set("key1", "value1")
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    value, _ := store.Get("key1")
    if value != "value1" {
        t.Errorf("Expected value 'value1', got '%s'", value)
    }
}

func TestKVStoreImpl_Delete(t *testing.T) {
    store := NewKVStoreImpl()
    store.Set("key1", "value1")

    err := store.Delete("key1")
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    _, err = store.Get("key1")
    if err == nil {
        t.Error("Expected KeyNotFoundError, got nil")
    }
}

