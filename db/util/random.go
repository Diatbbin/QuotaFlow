package util

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomInt(min, max int64) int64 {
	return min + rng.Int63n(max-min+1)
}

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range n {
		result[i] = letters[rng.Intn(len(letters))]
	}
	return string(result)
}

func RandomOwner() string {
	return RandomString(8)
}	

func RandomBalance() int64 {
	return RandomInt(0, 1000)
}