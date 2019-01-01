package wxlogin

import (
	"testing"
)

// func TestIsExpired(t *testing.T) {
// 	accExpired := OAuthAccess{
// 		createdAt: time.Now().AddDate(0, 0, -30),
// 		updatedAt: time.Now().Add(-time.Second * 7200),
// 	}

// 	t.Logf("Is access token expired: %t\n", accExpired.IsAccessExpired())
// 	t.Logf("Is refresh token expired: %t\n", accExpired.IsRefreshExpired())
// }

func TestMD5(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()

	t.Logf("Session ID: %s\n", acc.SessionID)
}

func TestEmbeddedStruct(t *testing.T) {
	acc := OAuthAccess{}

	t.Logf("OAuth Access: %+v\n", acc)
}

func TestSaveAccess(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()
	t.Logf("OAuth Access: %+v\n", acc)

	err := devEnv.SaveAccess(mockWxApp.AppID, acc, mockClient)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Saved OAuth Access: %+v\n", acc)
}
func TestLoadAccess(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()

	err := devEnv.SaveAccess(mockWxApp.AppID, acc, mockClient)

	t.Logf("Session ID for OAuth access: %s", acc.SessionID)

	if err != nil {
		t.Error(err)
		return
	}

	dbAcc, err := devEnv.LoadAccess(mockWxApp.AppID, acc.SessionID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Loaded access: %+v\n", dbAcc)
}

func TestUpdateAccess(t *testing.T) {
	wxID := newMockWxID()
	acc := wxID.generateAccess()

	err := devEnv.SaveAccess(mockWxApp.AppID, acc, mockClient)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Original access: %+v\n", acc)

	accessToken := generateToken()
	err = devEnv.UpdateAccess(acc.SessionID, accessToken)

	if err != nil {
		t.Error(err)
	}

	t.Logf("New access token: %+s\n", accessToken)
}
