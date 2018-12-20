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

func TestRandomBase64(t *testing.T) {
	t1, _ := RandomBase64(3)
	t.Logf("Base64url encoding 3 bytes: %s\n", t1)

	t2, _ := RandomBase64(6)
	t.Logf("Base64url encoding 6 bytes: %s\n", t2)

	t3, _ := RandomBase64(9)
	t.Logf("Base64url encoding 9 bytes: %s\n", t3)
}

func TestFakeWxCode(t *testing.T) {
	code, _ := RandomBase64(24)

	t.Logf("Wechat OAuth code: %s\n", code)
}

func TestFakeWxToken(t *testing.T) {
	token, _ := RandomBase64(82)

	t.Logf("Wechat token: %s\n", token)
}

func TestFakeWxUnionID(t *testing.T) {
	id, _ := RandomBase64(28)

	t.Logf("Wechat ID: %s\n", id)
}
