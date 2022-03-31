package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func InitUtils() {
	rand.Seed(time.Now().UnixNano())
}

// Returns an int >= min, < max
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(RandomInt(65, 90))
	}
	return string(bytes)
}

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
