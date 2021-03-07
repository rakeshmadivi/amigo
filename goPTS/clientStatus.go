package main

import (
	"fmt"
	//"log"
	//"time"
	//"strconv"
	//"bufio"
	//"os"
	//"strings"

	//guuid "github.com/google/uuid"
	"sync"
	"log"
	//"math/rand"
	//"time"
	//"net"
	"net/http"
	"encoding/json"
	//"strings"
	//"github.com/go-playground/validator"
	"io"	// for wrting strings to ResponseWriter
)

type TEST struct{
	name string
	args string
}

var wg sync.WaitGroup

var lck sync.Mutex
type MTX struct {
	mtx sync.Mutex
	cnt int
}

var mutex MTX

func LOCK(){
	mutex.mtx.Lock()
	mutex.cnt++
}

func UNLOCK(){
	mutex.mtx.Unlock()
	mutex.cnt--
}

func LockCount(){
	fmt.Println("--| LOCK COUNT:", mutex.cnt)
}

// Test Details
var Test TEST 

// Test Status
type SYNC_STATUS struct {
	groupSize, init, running, done, exit int
	SYNC_STATE string
}


var MAX_UID = 4

var syncStatus = make(map[string]*SYNC_STATUS) 

var SUT_status SUT 

// Variables to store info sent by Clients
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

// Following sutMap uses TestUID for indexing, Element maps uses srcIP for indexing.
var sutMap = make(map[string]map[string]*SUT)	// TODO adjust map keys to match type of first arg of SUT

type REPLY struct{
	Reply REQ
}

var ClientStates = map[string]int{"INIT":0,"RUNNING":1,"DONE":2,"EXIT_ACK":3}

func main(){

	/*
	fmt.Println("Enter IP:")
	fmt.Scanln(&ip)
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter STATUS: ")
	status, _ = reader.ReadString('\n')
	fmt.Println(status)

	updateStatus(ip,status)
	fmt.Println(ip,status)
*/
	//testCode()
	setupRoutes()
	//wg.Wait()

}

func testCode(){
	type TC struct {
		a int
		b string
	}

	m := make(map[string]*TC)

	m["a"] = &TC{1,"str"}
	fmt.Println("TestCode: ",m["a"].a, m["a"].b)

}

func setupRoutes(){

	http.HandleFunc("/setStatus", postCall)
	http.HandleFunc("/request", postCall)
	http.HandleFunc("/getStatus", getCall)

	port := "8300"

	log.Print("Listening on:", port)

	http.ListenAndServe(":" + port,nil)

}

func postCall(w http.ResponseWriter, r *http.Request){	//TODO  Adjust the src type later

	var curl_req REQ

	// CHECK if JSON data is sent by Client
	if r.Header.Get("Content-Type") != "" {
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

		// Validate Fields
		if msg, c := validateFields(curl_req); c > 0 {
			http.Error(w,msg,http.StatusBadRequest)
			return
		}

		//fmt.Printf("[FROM-CURL-CALL]: %+v\n", curl_req)
	}

	LOCK()
	src := curl_req.MyIP //r.RemoteAddr	// TODO Re-visit for proper way to get Remote IP, Remove :PORT part 

	// Initialize SUT_status from POST data

	if uid_status,ok := syncStatus[curl_req.TestUID]; ok {
		if  curl_req.Status == "INIT" &&  uid_status.SYNC_STATE == "EXIT" {

			msg := "UNEXPECTED SITUATION"
			http.Error(w,msg,http.StatusBadRequest)

			uid_status.SYNC_STATE = "WAIT"

			return
		}
	}

	sutInfo := SUT{src,curl_req}

	fmt.Println("[REQUEST]",sutInfo)

	lck.Lock()
	cStatus, _ := setSUTstatus(sutInfo)
	lck.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&cStatus) //reply)
	//fmt.Println("[Updated STATUS]",cStatus)

	UNLOCK()
}

func validateFields(r REQ) (string, int) {

	cnt := 0
	msg := "No value provided for required field: "

	if r.TestType == "" {
		msg += "TestType"
		cnt += 1
	}
	if r.GroupSize == 0 {
		msg += " GroupSize"
		cnt += 1
	}
	if r.TestUID == "" {
		msg += " TestUID"
		cnt += 1
	}
	if r.Test == "" {
		msg += " Test"
		cnt += 1
	}
	/*
	if r.Args == "" {
		msg += " Args"
		cnt += 1
	}
	*/
	if r.Status == "" {
		msg += " Status"
		cnt += 1
	}

	return msg, cnt
}

func getIP(r *http.Request) {
	realIP := r.Header.Get("X-REAL-IP")
	forwardIP := r.Header.Get("X-FPRWARDED-FOR")
	//remoteIP := strings.Split(r.RemoteAddr,":")[2]
	remoteIP := r.RemoteAddr

	fmt.Println("Real:",realIP,"Forwarded:",forwardIP,"Remote:",remoteIP)
}

func getCall(w http.ResponseWriter, r *http.Request){
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

	fmt.Printf("STATUS of: %+v\n", get.Testuid)

	LOCK()

	/*
	if mutex.cnt != 0 {
		msg := "CURRENT-LOCK-COUNT:" + strconv.Itoa(mutex.cnt) +"Exiting...."

		http.Error(w,msg,http.StatusBadRequest)

		return
	}
	*/

	status := getStatus(get.Testuid)

	/*
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&status) //reply)
	*/

	io.WriteString(w,status+"\n")

	UNLOCK()
}

func setSUTstatus(sut SUT) (string,bool) { //, wg *sync.WaitGroup){

	src := sut.src
	test := sut.test
	uid := test.TestUID

	// Print Entries
	//PrintEntries()

	if _,ok := sutMap[uid] ; !ok && len(sutMap) > MAX_UID {
		fmt.Println("Exceeded Max Map Size..! Skipping Operation.")
		return "Exceeded MaxUIDs for New UID: "+uid,false
	}

	//fmt.Println("[REQUEST]",sut)

	// Check if valid state
	_,valid_state := ClientStates[test.Status]
	if !valid_state {
		fmt.Println("Invalid Client State!")

		return "INVALID STATE",false
	}

	_,uid_in_sutMap := sutMap[uid]
	if !uid_in_sutMap {
		sutMap[uid] = make(map[string]*SUT)
	}

	if k, ok := sutMap[uid][src]; ok {
		if  test.Status == k.test.Status {
			msg := "[ "+src+"/"+test.TestUID+"/"+test.Status+" ] IP/TestUID/Status Already Present."
			return msg,true
		}
	}else{
		fmt.Println("[New SUT Entry]")
	}

	sutMap[uid][src] = &SUT{src,test}
	//fmt.Println("[ Updated SUT ]", uid,src,":",sutMap[uid][src])
/*
	k,uid_in_syncStatus := syncStatus[uid]
	if !uid_in_syncStatus {
		fmt.Println("[New STATUS Entry] ",uid)

		tmp := SYNC_STATUS{}

		tmp.groupSize = test.GroupSize

		if test.Status == "INIT" {
			tmp.init = 1
		}else if test.Status == "RUNNING" {
			tmp.running = 1
		}else if test.Status == "DONE" {
			tmp.done = 1
		}else if test.Status == "EXIT_ACK" {
			tmp.exit = 1
		}

		tmp.SYNC_STATE = "WAIT"

		syncStatus[uid] = &tmp

		return "Added New Status Entry for: "+uid,true
	}else {

		if test.Status == "INIT" &&  k.SYNC_STATE == "EXIT" {
			return "Unexpected condition.",false
		}

		// UID Present in syncStatus
		if test.GroupSize != k.groupSize {
			fmt.Println("Group Size mismatch!")
			return "GroupSize for TestUID: "+uid+" Mismatched. Expected Size: " + strconv.Itoa(k.groupSize), false
		}
		if k.init > k.groupSize || k.running > k.groupSize || k.done > k.groupSize {
			fmt.Println("[ ",k," ] Sync Status (init/running/done) Count > group Size !")
			return "Exceeded GroupSize", false
		}


		if test.Status == "INIT" {
			k.init += 1
			if k.init == k.groupSize && k.running == 0 && k.done == 0 {
				k.SYNC_STATE = "START"
				//k.init = 0
			}
		}else if test.Status == "RUNNING" {
			k.running += 1
			k.init -= 1
			if k.running == k.groupSize && k.init == 0 && k.done == 0 {
				k.SYNC_STATE = "RUNNING"
				//k.running = 0
			}
		}else if test.Status == "DONE" {
			k.done += 1
			k.running -= 1
			if k.done == k.groupSize && k.running == 0 && k.init == 0 {
				k.SYNC_STATE = "EXIT"
				//k.done = 0
			}
		}else if test.Status == "EXIT_ACK" {
			k.exit += 1
			k.done -= 1
			if k.exit == k.groupSize { //&& k.done == 0 && k.running == 0 && k.init == 0 {
				// RESET SUT and STATUS entries

				fmt.Println("Got EXIT_ACK from all: ",k)

				k.SYNC_STATE = "WAIT"
				k.init = 0
				k.running = 0
				k.done = 0
				k.exit = 0

				// Delete entry in SUT and STATUS
				fmt.Println("Deleting entry of TestUID: "+uid)
				delete(sutMap,uid)
				delete(syncStatus,uid)

				PrintEntries()
				return "Deleted: "+uid, true
			}
		}
		return "Updated Status Entry for: "+uid,true
	}
	*/
	PrintEntries()
	return "Updated SUT for TestUID: "+uid+src,true
}

func getStatus(uid string) string { //wg *sync.WaitGroup) string {
/*
	lck.Lock()
	status, ok := syncStatus[uid]
	if !ok {

		msg := "No Status for TestUID:"+uid

		PrintEntries()

		return msg
	}

	reply := status.SYNC_STATE
	fmt.Println("[ REPLY for TestUID -",uid," ] ", reply)
	lck.Unlock()
	return reply
	*/

	return getReply(uid)
}

var next = "INIT"
func getReply(uid string) string {
	lck.Lock()
	var sREPLY = "WAIT"
	
	init,running,done,exit := 0,0,0,0
	
	var g_size = 0

	k,ok := sutMap[uid]
	if ok {
		for _,v := range k {

			g_size = v.test.GroupSize

			if v.test.Status == "INIT" {
				init += 1
			}else if v.test.Status == "RUNNING" {
				running += 1
			}else if v.test.Status == "DONE" {
				done += 1
			}else if v.test.Status == "EXIT_ACK" {
				exit += 1
			}
		}

		if init == g_size {
			sREPLY = "START"
			next = "RUNNING"
		}
		if running == g_size && init == 0 {
			sREPLY = "RUNNING"
			next = "DONE"
		} 
		if done == g_size && running == 0 {
			sREPLY = "EXIT"
			next = "EXIT_ACK"
		}
		if exit == g_size && done == 0 {
			sREPLY = "WAIT"
			next = "INIT"
		}

		fmt.Println("Current Status:",init,running,done,exit," /",g_size," Replying:",sREPLY)
	}
	lck.Unlock()
	return sREPLY
}

func PrintEntries(){
	// SUT Entries
	if len(sutMap) == 0 {
		fmt.Println("No SUT Entries Found!")
	}else{
		fmt.Println("[Current SUT Entries - Map[src] : SUT{src,REQ}]")
		for _,v := range sutMap {
			//fmt.Println(v)
			for _,r := range v{
				fmt.Println(r)
			}
		}

	}
/*
	// STATUS Entries
	if len(syncStatus) == 0 {
		fmt.Println("No STATUS Entries Found!")
	}else{
		fmt.Println("[Current SUT Entries - TestUID : Status Entry]")
		for k,v := range syncStatus {
			fmt.Println(k,v)
		}
	}
	*/
}
