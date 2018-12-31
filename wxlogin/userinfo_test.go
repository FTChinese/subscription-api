package wxlogin

import "testing"

func TestSaveUserInfo(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()
	user := wxID.generateUserInfo()

	err := saveOneLogin(acc, user)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Saved userinfo: %+v\n", user)
}

func TestUpdateUserInfo(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()
	user := wxID.generateUserInfo()

	err := saveOneLogin(acc, user)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Create a wechat user: %s, %s\n", user.UnionID, user.NickName)

	newInfo := wxID.generateUserInfo()

	err = devEnv.UpdateUserInfo(newInfo)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Updated wechat user: %s, %s\n", newInfo.UnionID, newInfo.NickName)
}

// Generate multiple OAuthAccess with same OpenID,
// and multipe UserInfo with same OpenID and UnionID.
func TestMultiLogin(t *testing.T) {
	wxID := newMockWxID()

	oas := wxID.generateMultiAccess()
	user := wxID.generateUserInfo()

	for i := 0; i < len(oas); i++ {
		acc := oas[i]
		err := saveOneLogin(acc, user)
		if err != nil {
			t.Error(err)
			continue
		}

		t.Logf("Access token %s refresh token %s\n", acc.AccessToken, acc.RefreshToken)

		t.Logf("Saved userinfo for union id %s\n", user.UnionID)
	}
}
