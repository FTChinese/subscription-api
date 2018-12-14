package wxlogin

import "testing"

func TestSaveUserInfo(t *testing.T) {
	err := devEnv.SaveUserInfo(mockUserInfo, mockReqClient)

	if err != nil {
		t.Error(err)
	}
}
