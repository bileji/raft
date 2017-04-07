# raft

```
package main

import (
    "maple/raft"
    "time"
)

func main() {
    
    R1 := raft.New(0x0, 4, 1, 2, []uint64{0x0, 0x1})
    R2 := raft.New(0x1, 4, 1, 2, []uint64{0x0, 0x1})
    
    R1.WatchVoteFunc(func(id, leader, term uint64) bool {
        return R2.ReceiveVoteFunc(leader, term)
    })
    
    R1.WatchSyncFunc(func(term, leader uint64, nodes []uint64) error {
        R2.ReceiveSyncFunc(leader, term, nodes)
        return nil
    })
    
    R2.WatchVoteFunc(func(id, leader, term uint64) bool {
        return R1.ReceiveVoteFunc(leader, term)
    })
    
    R2.WatchSyncFunc(func(term, leader uint64, nodes []uint64) error {
        R1.ReceiveSyncFunc(leader, term, nodes)
        return nil
    })
    
    go R1.Introspect()
    go R2.Introspect()
    
    go func() {
        ticker := time.NewTicker(time.Duration(1 * time.Second))
        for {
            select {
            case <-ticker.C:
                R1.GetInfo()
                R2.GetInfo()
            }
        }
    }()
    
    for {
        select {}
    }
}
```