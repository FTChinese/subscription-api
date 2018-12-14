package util

import (
	"crypto/rand"
	"fmt"
)

func randomBytes(len int) ([]byte, error) {
	b := make([]byte, len)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// RandomHex generates a random hexadecimal number of 2*len chars
func RandomHex(len int) (string, error) {

	b, err := randomBytes(len)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
