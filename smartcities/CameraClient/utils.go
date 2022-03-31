package main

import (
	"math/rand"
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
