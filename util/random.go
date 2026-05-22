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

func RandomTool() string {
	tools := []string{"cursor", "copilot", "chatgpt"}
	return tools[rng.Intn(len(tools))]
}

func RandomTokenUsed() int64 {
	return RandomInt(0, 1000)
}

func RandomTokenLimit() int64 {
	return RandomInt(1000, 10000)
}
