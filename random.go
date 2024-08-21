package gotk

import (
	"strings"
	"time"

	"math/rand"
)

var (
	characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321"
	tkRand     = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func RandomInt(min, max int64) int64 {
	return min + tkRand.Int63n(max-min+1)
}

func RandomString(num int) string {
	var sb strings.Builder
	var size = len(characters)
	for i := 0; i < num; i++ {
		s := characters[tkRand.Intn(size)]
		sb.WriteByte(s)
	}
	return sb.String()
}
