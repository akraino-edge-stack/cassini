package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"time"

	"go.uber.org/zap"
)

type Node struct {
	name     string // name in cluster
	pass     bool   // if verify passed
	ip       string // node ip address
	data     string // random data for verify node
	waitTime int64  // time to start check agent avail
}

// hold all node
var clusterNodes map[string]*Node

func InitNodes() {
	clusterNodes = make(map[string]*Node)
}

func GetNewNode() *Node {
	// $kubectl get nodes -o json
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}

	var values map[string]interface{}
	err = json.Unmarshal(output, &values)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}

	items := reflect.ValueOf(values["items"]) // Array
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Interface()
		mapItem, _ := item.(map[string]interface{})

		metadata := mapItem["metadata"].(map[string]interface{})
		name := metadata["name"].(string)

		annotations := metadata["annotations"].(map[string]interface{})
		ip := annotations["k3s.io/internal-ip"].(string)

		labels := metadata["labels"].(map[string]interface{})
		isMaster := labels["node-role.kubernetes.io/master"]

		statusdata := mapItem["status"].(map[string]interface{})
		conddata := statusdata["conditions"].([]interface{})
		isReady := false
		for _, cond := range conddata {
			cdmap := cond.(map[string]interface{})
			if cdmap["type"].(string) == "Ready" {
				isReady = cdmap["status"].(string) == "True"
				break
			}
		}

		if false == isReady {
			zap.L().Warn("Node is NOT Ready:" + ip)
			continue // not online node, skip the verify process
		}

		// check if cached
		node, ok := clusterNodes[name]
		if ok {
			if node.pass {
				continue // varified
			}
			// 300 second time out after first ping. the node is still "Ready"
			if (time.Now().Unix() - node.waitTime) > 300 {
				zap.L().Warn("Node is Ready and ping timeout:" + node.ip)
				return node // this node need to be verify
			} else {
				continue // wait time out
			}
		}

		// new in cache
		node = &Node{
			name:     name,
			pass:     false,
			ip:       ip,
			data:     "",
			waitTime: 0,
		}
		clusterNodes[name] = node

		if isMaster == "true" {
			node.pass = true
			continue // master no need to varify
		} else {
			node.data = RandomString(32)
		}

		// if not online
		if false == node.Ping() {
			zap.L().Warn("Node is Ready but CAN NOT ping:" + node.ip)
			node.waitTime = time.Now().Unix()
			continue // k3s Ready status may refresh delayed, use timeout to recheck it
		}

		return node // node is online. need be varified
	}

	return nil
}

func GetNodeByIp(ip string) *Node {
	for _, node := range clusterNodes {
		if node.ip == ip {
			return node
		}
	}
	return nil
}

func (node *Node) Ping() bool {
	cmd := exec.Command("ping", node.ip, "-c", "4", "-W", "5")
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (node *Node) IsAgentReady() bool {
	url := fmt.Sprintf("http://%s:8301/version", node.ip)
	code, _ := CurlString("GET", url, "")
	return code == http.StatusOK
}

func (node *Node) RequestVerify() bool {
	url := fmt.Sprintf("http://%s:8301/encrypt", node.ip)
	code, encStr := CurlString("POST", url, node.data)
	if code != http.StatusOK {
		return false
	}

	// make sure client was created
	CurlString("POST", "http://127.0.0.1:8300/client", "{\"Name\": \"GoClient\"}")

	/* Send data to Agent, agent use public key to encrypt the data.
	EncryptDate send back to Server, server decode it, if ok, verify pass.
	NOTE: Server create key pairs and export public key
	Manually create sign key in SERVER:
	curl -v -d '{"Name": "GoClient", "KeyName": "MyEncKey"}' 127.0.0.1:8300/keyenc
	Manually export sign key's pub in SERVER:
	curl -v -X GET -d '{"Name": "GoClient", "KeyName": "MyEncKey"}' 127.0.0.1:8300/key
	Manually import sign pub key in AGNET:
	curl -v -d '{"Name": "GoClient", "KeyName": "MyPubKey", "Message":"ssh-rsa MIIBCgKCAQEA2OB/QQQfFdMEe/SFmIYSRWbLCstF9F6lLlV79FUW5iDoDxUpTp6upA97d5CdvHsMeTjXANBu4jMTs2HHl1pCZNfictoruNk8wz7fpNHzWJBMxjBwi+yrL/WmXo/U/f3ed22nJ2ON3mjDJJczeCUdHILfOdVJdIolZ0acGImY3z6eBpnUraV4t27AVzKNh9QCjb/YH5AfKfZop7maE/mxroU/Ob/cARPXmxZ9srp2ZcFg3S0k8vq6IFvd5HL8p67D+yQUKRZGQvI/Gawzx5mMRY9fC0rb0O7gHsn+5tr/XQT6/jkYr21U1tV6Rxe59/3WOaVLKObbvZTlAkPzMqASwwIDAQAB GoClient_MyEncKey"}' 127.0.0.1:8300/key
	*/
	param := fmt.Sprintf("{\"Name\": \"GoClient\", \"KeyName\": \"MyEncKey\", \"Message\": \"%s\"}", encStr)
	code, plainText := CurlString("POST", "http://127.0.0.1:8300/decrypt", param)
	if code == http.StatusOK && plainText == node.data {
		node.pass = true // update cache
	}
	return node.pass
}

func (node *Node) RemoveSelf(reason string) {
	zap.L().Error("Verify fail:" + node.ip + " reason:" + reason)
	zap.L().Info("Removing node from k3s server...")

	// $kubectl cordon nodename
	cmd := exec.Command("kubectl", "cordon", node.name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	// $kubectl drain nodename --delete-local-data --force --ignore-daemonsets
	cmd = exec.Command("kubectl", "drain", node.name, "--delete-local-data", "--force", "--ignore-daemonsets")
	_, err = cmd.CombinedOutput()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	// $kubectl delete node nodename
	cmd = exec.Command("kubectl", "delete", "node", node.name)
	_, err = cmd.CombinedOutput()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	zap.L().Info("Remove done")
}
