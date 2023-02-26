package utils

import (
	"math/rand"
	"time"
)

// GetRandomInt returns a random integer between min and max (both included)
func GetRandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
