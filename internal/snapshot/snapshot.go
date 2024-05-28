// snapshot/snapshot.go
package snapshot

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
    "sync"
)

type Snapshot struct {
    mu       sync.RWMutex
    dir      string
    maxIndex int
}

func NewSnapshot(dir string) *Snapshot {
    return &Snapshot{
        dir: dir,
    }
}

func (s *Snapshot) Save(index int, data []byte) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if index <= s.maxIndex {
        return fmt.Errorf("snapshot index %d is not greater than the latest snapshot index %d", index, s.maxIndex)
    }

    snapshotPath := filepath.Join(s.dir, fmt.Sprintf("snapshot-%d.dat", index))
    err := ioutil.WriteFile(snapshotPath, data, 0644)
    if err != nil {
        return err
    }

    s.maxIndex = index
    return nil
}

func (s *Snapshot) Load(index int) ([]byte, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    snapshotPath := filepath.Join(s.dir, fmt.Sprintf("snapshot-%d.dat", index))
    data, err := ioutil.ReadFile(snapshotPath)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func (s *Snapshot) ListSnapshots() ([]int, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    files, err := ioutil.ReadDir(s.dir)
    if err != nil {
        return nil, err
    }

    var snapshots []int
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        name := file.Name()
        if ext := filepath.Ext(name); ext != ".dat" {
            continue
        }
        indexStr := name[len("snapshot-") : len(name)-len(".dat")]
        index, err := strconv.Atoi(indexStr)
        if err != nil {
            continue
        }
        snapshots = append(snapshots, index)
    }

    return snapshots, nil
}