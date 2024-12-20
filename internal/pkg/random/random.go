package random

import (
	"math/rand/v2"
	"unicode"
)

func RandomString(length int) string {
	str := make([]rune, length)
	for i := 0; i < length; i++ {
		s := rune(rand.Int32N(127))
		if unicode.IsLetter(s) {
			str[i] = s
			continue
		}
		i--
	}
	return string(str)
}
