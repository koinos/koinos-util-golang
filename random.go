package util

import "math/rand"

// GenerateBase58ID generates a random seed string
func GenerateBase58ID(length int) string {
	// Use the base-58 character set
	var runes = []rune("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

	// Randomly choose up to the given length
	seed := make([]rune, length)
	for i := 0; i < length; i++ {
		seed[i] = runes[rand.Intn(len(runes))]
	}

	return string(seed)
}
