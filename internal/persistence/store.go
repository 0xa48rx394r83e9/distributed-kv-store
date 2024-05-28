// persistence/store.go
package persistence

import (
    "bytes"
    "encoding/gob"
    "kvstore"
    "sync"
)

type PersistentStore struct {
    mu    sync.RWMutex
    data  map[string]string
    dirty bool
}

func NewPersistentStore() *PersistentStore {
    return &PersistentStore{
        data: make(map[string]string),
    }
}

func (s *PersistentStore) Get(key string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, ok := s.data[key]
    if !ok {
        return "", kvstore.KeyNotFoundError(key)
    }
    return value, nil
}

func (s *PersistentStore) Set(key, value string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
    s.dirty = true
    return nil
}

func (s *PersistentStore) Delete(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.data, key)
    s.dirty = true
    return nil
}

func (s *PersistentStore) IsDirty() bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.dirty
}

func (s *PersistentStore) MarkClean() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.dirty = false
}

func (s *PersistentStore) Snapshot() ([]byte, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(s.data)
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

func (s *PersistentStore) Restore(data []byte) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    var decodedData map[string]string
    dec := gob.NewDecoder(bytes.NewReader(data))
    err := dec.Decode(&decodedData)
    if err != nil {
        return err
    }

    s.data = decodedData
    s.dirty = false

    return nil
}