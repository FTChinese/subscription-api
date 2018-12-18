package wxlogin

import "testing"

func TestLoadAccountByWx(t *testing.T) {
	acnt, err := devEnv.FindAccountByWx(mockUserInfo.UnionID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Account: %+v\n", acnt)
	t.Logf("Wx account: %+v\n", acnt.Wechat)
}

func TestLoadAccountByFTC(t *testing.T) {
	acnt, err := devEnv.FindAccountByFTC(mockUserID)

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

// func TestSimpleQuery(t *testing.T) {
// 	fChan := make(chan FTCAccount)
// 	wChan := make(chan Wechat)
// 	mChan := make(chan Membership)

// 	go devEnv.FindAccountByFTC(mockUserID, fChan)
// 	go devEnv.FindWxAccount(mockUserInfo.UnionID, wChan)
// 	go devEnv.FindMemberByFTC(mockUserID, mChan)

// 	acnt, wx, member := <-fChan, <-wChan, <-mChan

// 	t.Logf("%+v\n", acnt)
// 	t.Logf("%+v\n", wx)
// 	t.Logf("%+v\n", member)
// }
