package main

import (
	"fmt"
	//"log"
	//"time"
	"strconv"
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
)

type TEST struct{
	name string
	args string
}

var wg sync.WaitGroup

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
	groupSize, init, running, done int
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

var ClientStates = map[string]int{"INIT":0,"RUNNING":1,"DONE":2}

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

		fmt.Printf("[FROM-CURL-CALL]: %+v\n", curl_req)
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

	fmt.Println("[SET-Request]",sutInfo)

	cStatus, _ := setSUTstatus(sutInfo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&cStatus) //reply)
	fmt.Println("cStatus:",cStatus)
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

	fmt.Printf("RECEIVED: %+v\n", get.Testuid)

	LOCK()

	LockCount()

	if mutex.cnt != 0 {
		msg := "CURRENT-LOCK-COUNT:" + strconv.Itoa(mutex.cnt) +"Exiting...."

		http.Error(w,msg,http.StatusBadRequest)

		return
	}

	status := getStatus(get.Testuid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&status) //reply)
	fmt.Println("Replied:", status)
	UNLOCK()
}

func setSUTstatus(sut SUT) (SYNC_STATUS,bool) { //, wg *sync.WaitGroup){

	src := sut.src
	test := sut.test
	uid := test.TestUID

	fmt.Println("SUT MAP:")
	for k,v := range sutMap {
		fmt.Println("TestUID:",k,"Has",len(v),"Elements")
	}

	fmt.Println("SYNC STATUS:")
	for k,v := range syncStatus {
		fmt.Println("TestUID:",k,"STATUS:",*v)
	}

	if _,ok := sutMap[uid] ; !ok && len(sutMap) > MAX_UID {
		fmt.Println("Exceeded Max Map Size..! Skipping Operation.")
		return SYNC_STATUS{SYNC_STATE:"Exceeded MaxUIDs for New UID: "+uid},false
	}

	fmt.Println("[REQUEST]",sut)

	// Check if valid state
	_,valid_state := ClientStates[test.Status]
	if !valid_state {
		fmt.Println("Invalid Client State!")

		//mutex.Unlock()
		return SYNC_STATUS{SYNC_STATE:"INVALID STATE"},false
	}

	_,uid_in_sutMap := sutMap[uid]
	if !uid_in_sutMap {
		sutMap[uid] = make(map[string]*SUT)
	}

	if k, ok := sutMap[uid][src]; ok {
		if  test.Status == k.test.Status {
			msg := "[ "+src+"/"+test.TestUID+"/"+test.Status+" ] IP/TestUID/Status Already Present."
			return SYNC_STATUS{SYNC_STATE:msg},true
		}
	}else{
		fmt.Println("Adding New Request")
	}

	sutMap[uid][src] = &SUT{src,test}
	fmt.Println("After Updating SUT[",uid,"][",src,"]:",sutMap[uid][src])

	k,uid_in_syncStatus := syncStatus[uid]
	if !uid_in_syncStatus {
		fmt.Println("No Staus for :",uid," Appending new status...")

		tmp := SYNC_STATUS{}

		tmp.groupSize = test.GroupSize

		if test.Status == "INIT" {
			tmp.init = 1
		}else if test.Status == "RUNNING" {
			tmp.running = 1
		}else if test.Status == "DONE" {
			tmp.done = 1
		}
		tmp.SYNC_STATE = "WAIT"

		syncStatus[uid] = &tmp


		//return *SYNC_STATUS{}
		//mutex.Unlock()
		return *syncStatus[uid],true
	}else {

		if test.Status == "INIT" &&  k.SYNC_STATE == "EXIT" {
			msg := "Unexpected condition."
			return SYNC_STATUS{SYNC_STATE:msg},false
		}

		// UID Present in syncStatus
		if test.GroupSize != k.groupSize {
			fmt.Println("Group Size mismatch!")
			return SYNC_STATUS{SYNC_STATE:"GroupSize for TestUID: "+uid+" Mismatched. Expected Size: " + strconv.Itoa(k.groupSize) },false
		}
		if k.init > k.groupSize || k.running > k.groupSize || k.done > k.groupSize {
			fmt.Println("[ ",k," ] Sync Status (init/running/done) Count > group Size !")
			return SYNC_STATUS{SYNC_STATE:"Exceeded GroupSize"},false
		}


		if test.Status == "INIT" {
			k.init += 1
			if k.init == k.groupSize && k.running == 0 && k.done == 0 {
				k.SYNC_STATE = "START"
			}
		}else if test.Status == "RUNNING" {
			k.running += 1
			k.init -= 1
			if k.running == k.groupSize && k.init == 0 && k.done == 0 {
				k.SYNC_STATE = "RUNNING"
			}
		}else if test.Status == "DONE" {
			k.done += 1
			k.running -= 1
			if k.done == k.groupSize && k.running == 0 && k.init == 0 {
				k.SYNC_STATE = "EXIT"
			}
		}
		fmt.Println(syncStatus[uid])
		return *syncStatus[uid],true
	}
}

func getStatus(uid string) string { //wg *sync.WaitGroup) string {

	status, ok := syncStatus[uid]
	if !ok {

		msg := "No Status for TestUID:"+uid

		fmt.Println("Existing Requests:", sutMap)
		fmt.Println("Existing Statuses:", syncStatus)

		return msg
	}

	log.Println("[ CURRENT STATE ] ", status.SYNC_STATE)
	return status.SYNC_STATE
}

