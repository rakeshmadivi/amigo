package main

import (
	"fmt"
	"sync"
)

type RWMap struct {
	sync.RWMutex
	m map[string]int
}

// Get is a wrapper for getting the value from the underlying map
func (r RWMap) Get(key string) int {
	r.RLock()
	defer r.RUnlock()
	return r.m[key]
}

// Set is a wrapper for setting the value of a key in the underlying map
func (r RWMap) Set(key string, val int) {
	r.Lock()
	defer r.Unlock()
	r.m[key] = val
}

// Inc increases the value in the RWMap for a key.
//   This is more pleasant than r.Set(key, r.Get(key)++)
func (r RWMap) Inc(key string) {
	r.Lock()
	defer r.Unlock()
	r.m[key]++
}

func main() {

	// Init
	counter := RWMap{m: make(map[string]int)}

	// Get a Read Lock
	counter.RLock()
	_,ok := counter.m["Key"]
	if !ok {fmt.Println("No key found.")}
	counter.RUnlock()

	// the above could be replaced with
	_ = counter.Get("Key")

	// Get a write Lock
	counter.Lock()
	counter.m["some_key"]++
	counter.Unlock()

	// above would need to be written as 
	counter.Inc("some_key")
}
