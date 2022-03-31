package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// method = "GET", "POST", "PUT", "DELETE"
func CurlString(method string, url string, data string) (int, string) {
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		return http.StatusBadRequest, ""
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: time.Second * time.Duration(10)}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusBadRequest, ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, ""
	}
	return resp.StatusCode, string(body)
}
