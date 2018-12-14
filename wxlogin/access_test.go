package wxlogin

import "testing"

func TestSaveAccess(t *testing.T) {
	err := devEnv.SaveAccess(mockAccess, mockReqClient)

	if err != nil {
		t.Error(err)
	}
}

func TestLoadAccess(t *testing.T) {
	acc, err := devEnv.LoadAccess(mockAccess.OpenID, mockReqClient)

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(acc)
}
