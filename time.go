package raft

import "time"

func iToSec(i int) time.Duration {
   return time.Duration(i) * time.Second
}