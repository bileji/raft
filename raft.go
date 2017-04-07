package raft

import (
    "time"
    "errors"
    "fmt"
    "math/rand"
)

type raft struct {
    ID              uint64
    
    Term            uint64
    
    leader          uint64
    
    TenureTimer     int
    HeartbeatTicker int
    TimeoutTimer    int
    
    role            chan uint32
    
    timeout         *time.Timer
    tenure          *time.Timer
    heartbeat       *time.Ticker
    
    refresh         chan struct{}
    
    nodes           rwLock
    
    ballotBox       rwLock
    
    //存根
    stub            struct {
                        vf   uint64
                        term uint64
                    }
    
    voteFunc        func(id, leader, term uint64) bool
    syncFunc        func(term, leader uint64, nodes []uint64) error
}

func (r *raft) civilian() {
    r.role <- CIVILIAN
}

func (r *raft) candidate() {
    r.role <- CANDIDATE
}

func (r *raft) president() {
    r.role <- PRESIDENT
}

func (r *raft) prepareCampaign() {
    r.catchErr()
    
    r.timeout = time.NewTimer(iToSec(r.TimeoutTimer))
    r.Term++
    
    if r.Term <= r.stub.term {
        r.Term++
    }
    
    r.stub.vf = r.ID
    r.stub.term = r.Term
    
    r.ballotBox.Reset().Set(r.ID, r.Term)
    
    nodes := r.nodes.GetKeys()
    
    addUpVoteFor := func(id uint64) {
        defer r.catchErr()
        if r.voteFunc != nil && r.voteFunc(id, r.ID, r.Term) {
            r.ballotBox.Set(id, r.Term)
        }
    }
    
    go func(members []uint64) {
        for _, id := range members {
            if id != r.ID {
                // todo runtime.Goexit()
                go addUpVoteFor(id)
            }
        }
    }(nodes)
}

func (r *raft) standby() {
    second := iToSec(r.TenureTimer)
    r.tenure = time.NewTimer(second)
    for {
        select {
        case <-r.refresh:
            r.tenure.Reset(iToSec(r.TenureTimer))
        case <-r.tenure.C:
            r.tenure.Stop()
            go r.candidate()
            return
        }
    }
}

func (r *raft) randWait(millisecond int) {
    numSec := rand.New(rand.NewSource(time.Now().UnixNano()))
    time.Sleep(time.Duration(numSec.Intn(millisecond)) * time.Millisecond)
}

func (r *raft) campaign() {
    again:r.prepareCampaign()
    for {
        select {
        case <-r.timeout.C:
            if r.aggregate() != true {
                // 随机休息一段时间
                r.randWait(1000)
                goto again
            }
            r.timeout.Stop()
            go r.president()
            return
        case <-r.refresh:
            r.timeout.Stop()
            go r.civilian()
            return
        }
    }
}

func (r *raft) order() {
    r.leader = r.ID
    second := iToSec(r.HeartbeatTicker)
    r.heartbeat = time.NewTicker(second)
    for {
        select {
        case <-r.refresh:
            r.heartbeat.Stop()
            go r.civilian()
            return
        case <-r.heartbeat.C:
            go func(term, leader uint64, nodes []uint64) {
                if r.syncFunc != nil {
                    defer r.catchErr()
                    r.syncFunc(term, leader, nodes)
                }
            }(r.Term, r.ID, r.nodes.GetKeys())
        }
    }
}

func (r *raft) aggregate() bool {
    return len(r.ballotBox.GetKeys()) >= len(r.nodes.GetKeys()) / 2 + 1
}

func (r *raft) dispatcher(role uint32) {
    switch role {
    case CIVILIAN:
        go r.standby()
    case CANDIDATE:
        go r.campaign()
    case PRESIDENT:
        go r.order()
    default:
        panic(errors.New("undefined identity"))
    }
}

func (r *raft) Introspect() {
    
    if len(r.role) < 1 {
        r.role <- CIVILIAN
    }
    
    for {
        select {
        case id := <-r.role:
            r.dispatcher(id)
        }
    }
}

func (r *raft) catchErr() {
    if err := recover(); err != nil {
        fmt.Println(err)
    }
}

func (r *raft) WatchVoteFunc(vote func(id, leader, term uint64) bool) {
    r.voteFunc = vote
}

func (r *raft) WatchSyncFunc(sync func(term, leader uint64, nodes []uint64) error) {
    r.syncFunc = sync
}

func (r *raft) ReceiveVoteFunc(id, term uint64) bool {
    if term <= r.stub.term {
        return false
    }
    
    r.refresh <- struct{}{}
    r.stub.vf = id
    r.stub.term = term
    
    return true
}

func (r *raft) ReceiveSyncFunc(id, term uint64, nodes []uint64) {
    if term >= r.Term {
        // 信号必须在前
        r.refresh <- struct{}{}
        
        r.Term = term
        r.leader = id
        r.nodes.BatchSet(term, nodes)
    }
}

func (r *raft) GetInfo() {
    fmt.Println("ID:", r.ID, "届数:", r.Term, "领导:", r.leader, "存根:", r.stub, "投票箱:", r.ballotBox.Member(), "子节点:", r.nodes.GetKeys())
}

/* todo 当tenure - timeout = heartbeat时易出现,
 * todo 竞选获胜同步信息还未来得及同步子节点就开始下一轮竞选的情况
 * todo 如此反复，陷入循环无法跳出
 */
func New(id uint64, tenure, heartbeat, timeout int, nodes []uint64) *raft {
    raft := &raft{
        role: make(chan uint32, 1),
        Term: 0x0,
        ID: id,
        refresh: make(chan struct{}),
        TenureTimer: tenure,
        HeartbeatTicker: heartbeat,
        TimeoutTimer: timeout,
    }
    
    raft.nodes.BatchSet(0x0, nodes)
    
    return raft
}