package main

import (
	"fmt"
	"sync"
)

type RWMap struct {
	sync.RWMutex
	stat map[string]*SYNC_STATUS
}

// Get is a wrapper for getting the value from the underlying map
func (r RWMap) Get(key string) int {
	r.RLock()
	defer r.RUnlock()
	return r.stat[key]
}

// Set is a wrapper for setting the value of a key in the underlying map
func (r RWMap) Set(key string, val *SYNC_STATUS) {
	r.Lock()
	defer r.Unlock()
	r.stat[key] = val
}

var statMap = RWMap{stat: make(map[string]*SYNC_STATUS)}

func UpdateSyncStatus(uid string, val *SYNC_STATUS){
	statMap.Set(uid,val)
}

func ReadSyncStatus(uid string){
	statMap.Get(uid)
}
