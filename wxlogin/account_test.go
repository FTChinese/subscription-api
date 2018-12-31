package wxlogin

import "testing"

func TestLoadAccountByWx(t *testing.T) {
	acnt, err := devEnv.FindAccountByWx(myUnionID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Account: %+v\n", acnt)
	t.Logf("Wx account: %+v\n", acnt.Wechat)
}

func TestLoadAccountByFTC(t *testing.T) {
	acnt, err := devEnv.FindAccountByFTC(mockUUID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Account: %+v\n", acnt)
	t.Logf("Wx account: %+v\n", acnt.Wechat)
}

func TestBindAccount(t *testing.T) {
	err := devEnv.BindAccount(mockUUID, myUnionID)

	if err != nil {
		t.Error(err)
	}
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
