package random

import (
	"math/rand/v2"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(length int) string {
	str := make([]byte, length)

	for i := 0; i < length; i++ {
		j := rand.IntN(len(letters))
		str[i] = letters[j]
	}

	return string(str)
}
