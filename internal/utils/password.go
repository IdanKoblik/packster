package utils

import (
	"strings"
	"math/rand/v2"
)

func GenerateMask() string {
	n := rand.N(18) + 5
	return strings.Repeat("*", n)
}
