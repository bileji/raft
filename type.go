package raft

import (
    "sync"
)

const (
    CIVILIAN uint32 = 1 << iota
    CANDIDATE
    PRESIDENT
)

type rwLock struct {
    sync.RWMutex
    groups map[uint64]uint64
}

func (rw *rwLock) Reset() *rwLock {
    rw.Lock()
    defer rw.Unlock()
    rw.groups = make(map[uint64]uint64)
    return rw
}

func (rw *rwLock) Set(no1, no2 uint64) {
    rw.Lock()
    defer rw.Unlock()
    if rw.groups == nil {
        rw.groups = make(map[uint64]uint64)
    }
    rw.groups[no1] = no2
}

func (rw *rwLock) BatchSet(term uint64, nodes []uint64) {
    rw.Lock()
    defer rw.Unlock()
    if rw.groups == nil {
        rw.groups = make(map[uint64]uint64)
    }
    rw.groups = make(map[uint64]uint64, len(nodes))
    for _, node := range nodes {
        rw.groups[node] = term
    }
}

func (rw *rwLock) Get(no uint64) uint64 {
    rw.Lock()
    defer rw.Unlock()
    if rw.groups == nil {
        rw.groups = make(map[uint64]uint64)
    }
    return rw.groups[no]
}

func (rw *rwLock) Member() map[uint64]uint64 {
    rw.Lock()
    defer rw.Unlock()
    if rw.groups == nil {
        rw.groups = make(map[uint64]uint64)
    }
    return rw.groups
}

func (rw *rwLock) GetKeys() []uint64 {
    rw.Lock()
    defer rw.Unlock()
    if rw.groups == nil {
        rw.groups = make(map[uint64]uint64)
    }
    var keys []uint64
    for key := range rw.groups {
        keys = append(keys, key)
    }
    return keys
}