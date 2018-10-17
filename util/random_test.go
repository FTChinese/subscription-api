package util

import "testing"

func TestFakeWxID(t *testing.T) {
	token, err := RandomHex(8)

	if err != nil {
		t.Error(err)
	}

	t.Log(token)
}

func TestFakeWxKey(t *testing.T) {
	token, _ := RandomHex(16)

	t.Log(token)
}
