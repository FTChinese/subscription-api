package wxlogin

import "testing"

func TestLoadAccount(t *testing.T) {
	acnt, err := devEnv.LoadAccountByWx(mockUserInfo.UnionID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Account: %+v\n", acnt)
	t.Logf("Wx account: %+v\n", acnt.Wechat)
}

func TestBindAccount(t *testing.T) {
	err := devEnv.BindAccount("de6442b5-2e54-41a5-a33b-52473861855c", mockUserInfo.UnionID)

	if err != nil {
		t.Error(err)
	}
}

func TestCheckFTCAccount(t *testing.T) {
	s, err := devEnv.CheckFTCAccount(mockUserID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("FTC Account status: %+v\n", s)
}

func TestCheckWxAccount(t *testing.T) {
	s, err := devEnv.CheckWxAccount(mockUserInfo.UnionID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Wx account status: %+v\n", s)
}
