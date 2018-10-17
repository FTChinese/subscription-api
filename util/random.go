package util

import (
	"crypto/rand"
	"fmt"
)

// RandomHex generates a random hexadecimal number of 2*len chars
func RandomHex(len int) (string, error) {
	b := make([]byte, len)

	_, err := rand.Read(b)

	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", b), nil
}
