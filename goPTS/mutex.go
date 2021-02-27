package main

import (
	"fmt"
	"sync"
	"strconv"

)
var wg sync.WaitGroup
var bal int = 0
var crit sync.Mutex

var ms = make(map[int]string)

func deposit(d int, wg *sync.WaitGroup){
	crit.Lock()
	fmt.Printf("\nCur.Balance: %v Depositing: %v",bal,d)
	bal += d
	fmt.Printf("\nNew Balanc: %v",bal)
	crit.Unlock()
	wg.Done()
}

func withdraw(d int, wg *sync.WaitGroup){
	crit.Lock()
	fmt.Printf("\nCur.Balance: %v Withdrawing: %v",bal,d)
	bal -= d
	fmt.Printf("\nNew Balance: %v",bal)
	crit.Unlock()
	wg.Done()
}


func bank_example(){
	wg.Add(2)
	go deposit(100, &wg)
	go withdraw(100, &wg)
	wg.Wait()
}

func multi_worker_example(){

	for i:=0; i<5;i++{
		wg.Add(1)
		go worker(i,&wg)
	}
	wg.Wait()

	for k,v := range ms{
		fmt.Println(k,v)
	}
}

func worker(id int,wg *sync.WaitGroup){
	fmt.Println("Worker:",id)
	ms[id]="I am: " + strconv.Itoa(id)
	wg.Done()
}

func main(){
	multi_worker_example()
}

