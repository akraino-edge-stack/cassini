package main

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

func init() {
	InitLog()
}

func main() {
	// init Parsec client with restful API
	for {
		// 1. new parsec client
		code, _ := CurlString("POST", "http://127.0.0.1:8300/client", "{\"Name\": \"GoClient\"}")
		if code != http.StatusOK {
			time.Sleep(time.Second * time.Duration(10))
			continue // wait 10 sec then retry again
		}
		zap.L().Info("Parsec init SUCCESS")
		break // finish parsec init
	}

	// we have one goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		StartGinApis()
		wg.Done()
	}()

	wg.Wait() // wait goroutine exit
}
