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
	//"os"
	"math/rand"
	"time"
)

type TEST struct{
	name string
	args string
}

type SUT struct{
	src int //string	// TODO Adjust to consider IP
	test TEST
	status string
}

var SIZE = 10

var wg sync.WaitGroup
var mtx sync.Mutex

// Test Details
var Test TEST 

// Test Status
var SYNC_STATE = "WAIT" //= make(chan string) //= STATUS{TEST{"NONE","NONE"},"NONE"}

var sutMap = make(map[int]SUT)	// TODO adjust map keys to match type of first arg of SUT

var SUT_status SUT //= SUT{0,Test,"INIT"}

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

	defer close(SUT_STATUS)
	postCall()
	wg.Wait()
	getCall()

}

func generatePostCalls(){
	for i:=0; i< SIZE; i++ {
		postCall()
		randwait := rand.Intn(5)
		time.Sleep(time.Duration(randwait) * time.Second)
	}

	//wg.Wait()

}

func postCall(){
	// Initialize Test from POST data..
	Test = TEST{"PROG","ARGS"}

	// Initialize SUT_status from POST data

	SUT_status = SUT{0,Test,"INIT"}

	wg.Add(1)
	go setSUTstatus(SUT_status) //, &wg)
	//go updateSyncStatus()
}

func getCall(){
	getStatus() //&wg)
}

var SUT_STATUS = make(chan SUT)

// write to channel
func setSUTstatus(sut SUT) { //, wg *sync.WaitGroup){

	defer wg.Done()

	log.Println("Writing:",sut.status)
	//SUT_STATUS <- sut
	sutMap[sut.src] = sut

	status := sut.status

	mtx.Lock()
	if status == "INIT" {
		init_cnt += 1
	}else if status == "RUNNING" {
		init_cnt += 1
	}else if status == "DONE" {
		init_cnt += 1
	}
	mtx.Unlock()

	fmt.Println("INIT:",init_cnt,"RUNNING:",running_cnt,"DONE:",done_cnt)

	LEN := len(sutMap)	// TODO set to match init size than len(), since len() will be always true
	log.Println("Total Clients:",LEN)

	if init_cnt == LEN {
		SYNC_STATE = "START"

	}else if running_cnt == LEN {
		SYNC_STATE = "RUNNING"

	}else if done_cnt == LEN {
		SYNC_STATE = "EXIT"

	}
}

var init_cnt,running_cnt,done_cnt = 0,0,0

// Read from channel
func updateSyncStatus() {
	defer wg.Done()

	_sut := <-SUT_STATUS
	log.Println("Read from Channel:",_sut)

	status := _sut.status

	if status == "INIT" {
		init_cnt += 1
	}else if status == "RUNNING" {
		init_cnt += 1
	}else if status == "DONE" {
		init_cnt += 1
	}

	sutMap[_sut.src] = _sut

	fmt.Println("INIT:",init_cnt,"RUNNING:",running_cnt,"DONE:",done_cnt)

	LEN := len(sutMap)	// TODO set to match init size than len(), since len() will be always true
	log.Println("Total Clients:",LEN)

	if init_cnt == LEN {
		SYNC_STATE = "START"

	}else if running_cnt == LEN {
		SYNC_STATE = "RUNNING"

	}else if done_cnt == LEN {
		SYNC_STATE = "EXIT"

	}else {
		SYNC_STATE = "WAIT"
	}
}

func getStatus() string { //wg *sync.WaitGroup) string {
	log.Println("Current State:",SYNC_STATE)
	return SYNC_STATE
}

