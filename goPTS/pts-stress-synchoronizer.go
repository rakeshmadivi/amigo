
package main

import (
	"fmt"
	"sync"
	"strconv"
	"math/rand"
	"time"
	"net/http"
	"log"
	"encoding/json"
	"io"
)

type REQ struct {
	TestType string `json: testtype validate:"required"`	// ISOL/STRESS/NOISY ?
	GroupSize int `json: groupsize validate:"required"`
	TestUID string `json: testuid validate:"required"`//,omitempty`	// ID to uniquely identify this test
	Test string `json:test validate:"required"`//,omitempty`		// Test Profile name
	Args string `json: args validate:"required"`//,omitempty`			// Arguments to the test profile
	Status string `json: status validate:"required"`//,omitempty`		// Current status of Test on SUT
	MyIP string `json: myip` // TODO remove this after PTS simulation
}

type SUT struct{
	src string	// TODO Adjust to consider IP
	test REQ
}

type COUNTER struct {
	init,running,done,exit int
	state string
}

type RWMap struct {
	sync.RWMutex
	m map[string]int

	// Counters
	counters map[string]*COUNTER	// TestUIDs as keys
	sutMap map[string]map[string]*SUT	// TestUIDs as for outer Map and SourceIP for Inner Map as Keys
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

//var c = RWMap{m: make(map[string]int)}
//var c = RWMap{m: map[string]int{"k1":0} }
var c = RWMap{ m: map[string]int{"k1":0}, counters:make(map[string]*COUNTER) , sutMap : make(map[string]map[string]*SUT) }

var wg sync.WaitGroup
var enable_WG bool = false //true

func Reset(tid string)string{
	c.Lock()
	fmt.Println("Resetting Entries for TestUID:",tid)
	msg := "RESET COULD NOT PERFORM on " + tid
	if _,ok := c.sutMap[tid]; ok {
		delete(c.sutMap,tid)
		delete(c.counters,tid)

		_,ok1 := c.sutMap[tid]
		_,ok2 := c.counters[tid]

		if  !ok1 && !ok2 {
			msg = "SUCCECCFULLY RESET:" + tid
		}
	}

	c.Unlock()
	return msg
}

func Read(tid string) string {// tid string){
	//fmt.Println("Reading Value...")
	// Get a Read Lock
	c.RLock()
	/*
	_,ok := c.m["k1"]
	if !ok {fmt.Println("No key found.")}
	
	//fmt.Println("TID:"', tid, "Status:",c.counters[tid].state)
	fmt.Println("Counter:")
	for k,v := range c.counters {
		fmt.Println(k,v)
	}
	*/
	val,ok := c.counters[tid]

	status := ""
	if !ok {
		status = "WAIT"
	}else{
		status = val.state
	}

	fmt.Println("\n[tid -",tid,"] REPLY:",status)
	c.RUnlock()
	return status
}

func Write(req REQ){
	// Get a write Lock
	c.Lock()
	//c.m["k1"]++
	//fmt.Println("Written:", c.m["k1"])

	//tmp_str := strconv.Itoa(uid)
	tid := req.TestUID //"00-01" //tmp_str
	src := req.MyIP //"1.2.3."+tmp_str
	
	if _,ok := c.sutMap[tid]; !ok {
		c.sutMap[tid] = make(map[string]*SUT)
	}
	c.sutMap[tid][src] = &SUT{src,req} //REQ{"Normal",3,tid,"Test","Args","INIT",src}}
	fmt.Println("SUT Entries Size:",len(c.sutMap[tid]))

/*
	// Increment values accordingly
	status := c.sutMap[tid][src].test.Status

	if _,ok := c.counters[tid]; !ok {
		c.counters[tid] = &COUNTER{0,0,0,0,"WAIT"}
	}

	if status == "INIT" {
		c.counters[tid].init++

	}else if status == "RUNNING" {
		c.counters[tid].running++
		//c.counters[tid].init--

	}else if status == "DONE" {
		c.counters[tid].done++
		//c.counters[tid].running--

	}else if status == "EXIT_ACK" {
		c.counters[tid].exit++
		//c.counters[tid].done--
	}

	// Update State of the TID
	gsize := c.sutMap[tid][src].test.GroupSize
	i_cnt,r_cnt,d_cnt,e_cnt,status := c.counters[tid].init,c.counters[tid].running, c.counters[tid].done, c.counters[tid].exit, c.counters[tid].state
	if i_cnt == gsize {
		c.counters[tid].state = "START"

	}else if r_cnt == gsize {
		if (i_cnt - r_cnt) == 0 {
			c.counters[tid].state = "RUNNING"
			c.counters[tid].init = 0
		}else{
			c.counters[tid].state = "START"
		}

	}else if d_cnt == gsize {
		if (r_cnt - d_cnt) == 0 {
			c.counters[tid].state = "EXIT"
			c.counters[tid].running = 0
		}else{
			c.counters[tid].state = "RUNNING"
		}

	}else if e_cnt == gsize {
		if (d_cnt - e_cnt) == 0 {
			c.counters[tid].state = "WAIT"
			c.counters[tid].done = 0
			c.counters[tid].exit = 0
		}else{
			c.counters[tid].state = "EXIT"
		}

	}else{
		c.counters[tid].state = "WAIT"
	}
*/

	fmt.Println("[SUT Entres]:",c.sutMap[tid])
	c.Unlock()

	if enable_WG {
		wg.Done()
	}
}

func periodicStatusUpdate(){
	for {
		c.Lock()

		funcName :="[ UPDATER ]"
		fmt.Println(funcName,"sutMap Size:",len(c.sutMap),"Counter Size:",len(c.counters))

		for tid,v := range c.sutMap {

			fmt.Println(funcName,"TID:",tid,"Contains:",len(v))

			i_cnt,r_cnt,d_cnt,e_cnt := 0,0,0,0	//,"WAIT"
			gsize := 0
			for src,details := range v {
				fmt.Println(funcName,"SUT[",src,"] ",*details )
				status := details.test.Status
				gsize = details.test.GroupSize

				if status == "INIT" {
					i_cnt++

				}else if status == "RUNNING" {
					r_cnt++

				}else if status == "DONE" {
					d_cnt++

				}else if status == "EXIT_ACK" {
					e_cnt++
				}
			}

			fmt.Println(funcName,"TestID:",tid,"Counters: i,r,d,e = ",i_cnt,r_cnt,d_cnt,e_cnt)

			if _,ok := c.counters[tid]; !ok {
				c.counters[tid] = &COUNTER{}
			}

			// Update Count Values
			c.counters[tid].init = i_cnt
			c.counters[tid].running = r_cnt
			c.counters[tid].done = d_cnt
			c.counters[tid].exit = e_cnt

			if i_cnt == gsize {
				c.counters[tid].state = "START"

			}else if r_cnt == gsize {
				c.counters[tid].state = "RUNNING"

			}else if d_cnt == gsize {
				c.counters[tid].state = "EXIT"

			}else if e_cnt == gsize {
				c.counters[tid].state = "WAIT"
			}else{
				c.counters[tid].state = "WAIT"
			}
			fmt.Println("[UPDATED]:",c.counters[tid])
		}
		c.Unlock()
		time.Sleep(3*time.Second)
	}
}

func simulate(){
	states := []string{"INIT","RUNNING","DONE","EXIT_ACK"}
	
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	
	var str = ""
	for i :=0; i<3; i++ {
		wg.Add(2)
		idx := r1.Intn(len(states))
		state := states[idx]
		state = "INIT"
		req := REQ{"Normal",3,"00-01","Test","Args",state,"1.2.3."+strconv.Itoa(i)}
		go Write(req)
		go Read("00-01")
	}
	wg.Wait()
	fmt.Println(str)
}

func WriteStatus(w http.ResponseWriter, r *http.Request){
	var curl_req REQ
	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		msg := "Invalid Content Type! Only application/json content is accepted."
		log.Print(msg)
		http.Error(w,msg,http.StatusUnsupportedMediaType)
		return
	}

	// Parse POSTed JSON into REQ

	err := json.NewDecoder(r.Body).Decode(&curl_req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*/ Validate Fields
	if msg, c := validateFields(curl_req); c > 0 {
		http.Error(w,msg,http.StatusBadRequest)
		return
	}
	*/
	if enable_WG {
		wg.Add(1)
	}
	Write(curl_req)
	if enable_WG {
		wg.Wait()
	}
}

func ReadStatus(w http.ResponseWriter, r *http.Request){
	type cget struct { 
		Testuid string `json:testuid`
	}

	get := cget{}

	// CHECK if JSON data is sent by Client
	if r.Header.Get("Content-Type") != "application/json" {
		msg := "Invalid Content Type! Only application/json content is accepted."
		log.Print(msg)
		http.Error(w,msg,http.StatusUnsupportedMediaType)
		return
	}

	// Parse POSTed JSON into REQ

	err := json.NewDecoder(r.Body).Decode(&get)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := Read(get.Testuid)
	io.WriteString(w, reply+"\n")
}

func ResetStatus(w http.ResponseWriter, r *http.Request){
	type cget struct { 
		Testuid string `json:testuid`
	}

	get := cget{}

	// CHECK if JSON data is sent by Client
	if r.Header.Get("Content-Type") != "application/json" {
		msg := "Invalid Content Type! Only application/json content is accepted."
		log.Print(msg)
		http.Error(w,msg,http.StatusUnsupportedMediaType)
		return
	}

	// Parse POSTed JSON into REQ

	err := json.NewDecoder(r.Body).Decode(&get)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := Reset(get.Testuid)
	io.WriteString(w, reply+"\n")
}

func setupRoutes(){

	http.HandleFunc("/setStatus", WriteStatus)
	http.HandleFunc("/getStatus", ReadStatus)
	http.HandleFunc("/reset", ResetStatus)

	port := "8300"

	log.Print("Listening on:", port)

	http.ListenAndServe(":" + port,nil)

}

func main() {

	// Init

	// the above could be replaced with
	//_ = c.Get("Key")

	// above would need to be written as 
	//c.Inc("some_key")
	//simulate()
	go periodicStatusUpdate()
	setupRoutes()
}
