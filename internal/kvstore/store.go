// kvstore/store.go
package kvstore

import (
    "sync"
)

type KVStore interface {
    Get(key string) (string, error)
    Set(key, value string) error
    Delete(key string) error
}

type KVStoreImpl struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewKVStoreImpl() *KVStoreImpl {
    return &KVStoreImpl{
        data: make(map[string]string),
    }
}

func (s *KVStoreImpl) Get(key string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, ok := s.data[key]
    if !ok {
        return "", KeyNotFoundError(key)
    }
    return value, nil
}

func (s *KVStoreImpl) Set(key, value string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
    return nil
}

func (s *KVStoreImpl) Delete(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.data, key)
    return nil
}

type KeyNotFoundError string

func (e KeyNotFoundError) Error() string {
    return "key not found: " + string(e)
}

