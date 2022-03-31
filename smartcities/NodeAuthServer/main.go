package main

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

func init() {
	InitUtils()
	InitNodes()
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

	// we have two goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		// loop check new node in k3s
		for {
			node := GetNewNode()
			if node == nil {
				// check period
				time.Sleep(time.Duration(time.Second * 30))
				continue
			}
			zap.L().Info("Find new node to be verify:" + node.ip)
			// try get agent's restful version
			if false == node.IsAgentReady() {
				node.RemoveSelf("Agent Not Ready")
				continue
			}
			// start verify
			if false == node.RequestVerify() {
				node.RemoveSelf("Verify Fail")
				continue
			}
			zap.L().Info("verify pass:" + node.ip)
		}
		wg.Done()
	}()

	go func() {
		// use for edgfaas to get k3s server info
		StartGinApis()
		wg.Done()
	}()

	wg.Wait() // wait goroutine exit
}
