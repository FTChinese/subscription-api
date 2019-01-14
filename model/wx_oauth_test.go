package model

import "testing"

func TestSaveAccess(t *testing.T) {
	m := newMocker()
	acc := m.wxAccess()
	t.Logf("OAuth Access: %+v\n", acc)

	err := devEnv.SaveWxAccess(appID, acc, mockApp)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Saved OAuth Access: %+v\n", acc)
}

func TestLoadAccess(t *testing.T) {
	m := newMocker()

	acc, err := m.createWxAccess()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created access token: %+v\n", acc)

	dbAcc, err := devEnv.LoadWxAccess(appID, acc.SessionID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Loaded access: %+v\n", dbAcc)
}

func TestUpdateAccess(t *testing.T) {
	m := newMocker()

	acc, err := m.createWxAccess()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Original access: %+v\n", acc)

	accessToken := generateToken()
	err = devEnv.UpdateWxAccess(acc.SessionID, accessToken)

	if err != nil {
		t.Error(err)
	}
	t.Logf("New access token: %+s\n", accessToken)
}

func TestSaveUserInfo(t *testing.T) {
	m := newMocker()
	userInfo := m.wxUser()

	err := devEnv.SaveWxUser(userInfo)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateUserInfo(t *testing.T) {
	m := newMocker()

	userInfo, err := m.createWxUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Created wechat user: %+v\n", userInfo)

	newInfo := m.wxUser()
	err = devEnv.UpdateWxUser(newInfo)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Updated wechat user: %+v\n", newInfo)
}

// Generate multiple OAuthAccess with same OpenID,
// and multipe UserInfo with same OpenID and UnionID.
func TestMultiLogin(t *testing.T) {
	m := newMocker()

	for i := 0; i < 5; i++ {
		acc, err := m.createWxAccess()
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("A new login: %+v\n", acc)

		userInfo, err := m.createWxUser()
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("Save/Update userinfo: %+v\n", userInfo)
	}
}
