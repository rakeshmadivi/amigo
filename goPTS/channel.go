package main

import (
	"fmt"
	"sync"
)

var CH = make(chan string)
var wg sync.WaitGroup

func main(){

	defer close(CH)

	go write()
	wg.Add(1)
	go read()
	//msg := <-CH
	//fmt.Println("Read Message:",msg)
	wg.Wait()
	fmt.Println("DONE.")
}

func write(){
	CH <- "Hello from Write"
}

func read(){
	defer wg.Done()
	msg := <-CH
	fmt.Println("Read Message:",msg)
}
