package wxlogin

import "testing"

func TestGetAccessToken(t *testing.T) {
	acc, err := mockWxClient.GetAccessToken(mockCode)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Access token: %+v\n", acc)
}

func TestGetUserInfo(t *testing.T) {
	user, err := mockWxClient.GetUserInfo(mockAccess)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("User info: %+v\n", user)
}

func TestRefreshAccess(t *testing.T) {
	acc, err := mockWxClient.RefreshAccess(mockAccess.RefreshToken)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Access token: %+v\n", acc)
}

func TestIsValidAccess(t *testing.T) {
	ok := mockWxClient.IsValidAccess(mockAccess)

	t.Log(ok)
}
