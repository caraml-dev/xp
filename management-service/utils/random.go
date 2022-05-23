package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math"
)

func GenerateRandomBase16String(length int) (string, error) {
	buff := make([]byte, int(math.Ceil(float64(length)/2)))
	_, err := rand.Read(buff)
	if err != nil {
		return "", err
	}

	// Encode and strip one extra character where length is an odd number
	str := hex.EncodeToString(buff)[:length]
	return str, nil
}
