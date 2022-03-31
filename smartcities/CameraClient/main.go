package main

import (
	"sync"
)

func init() {
	InitUtils()
	InitLog()
}

func main() {
	// we have one goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		StartGinApis()
		wg.Done()
	}()

	wg.Wait() // wait goroutine exit
}
