package wxlogin

import (
	"testing"
)

// Group A: Code: 0, Message: , AccessToken: ***REMOVED***
// Group B: Code: 0, Message: , AccessToken: 17_0rLiFWTpiaWcjF6hflAa5cYDf9MqQDR1ZzsRDl9ddofzfp-u7QlL24PurFDL6ZohabLWogxjC4IzHOedb5B7AgR1NNwzE_HtIXQwxItpIvI, RefreshToken: 17_5wnUhIutNQBDJeyK-rxBCMMCP89SWG_mOzRqPUIWgPfmkoFgKninSKIySocpr_19Jp83sadiefG5_IALJ_XOIw5WpaOHZE1rBJuPTY-TnFs, OpenID: ob7fA0h69OO0sTLyQQpYc55iF_P0
func TestGetAccessToken(t *testing.T) {
	codeA := "001uvxwQ09nrU52byOxQ0AcowQ0uvxw3"
	codeB := "07183xW41z7iNS1eUcX41vNcW4183xWa"

	accA, err := mockWxApp.GetAccessToken(codeA)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Group A: %+v\n", accA)

	accB, err := mockWxApp.GetAccessToken(codeB)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Group B: %+v\n", accB)
}

func TestGetUserInfo(t *testing.T) {
	user, err := mockWxApp.GetUserInfo(myAccessToken, myOpenID)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("User info: %+v\n", user)
}

func TestRefreshAccess(t *testing.T) {
	acc, err := mockWxApp.RefreshAccess(myRefreshToken)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Access token: %#v\n", acc)
}

func TestIsValidAccess(t *testing.T) {
	ok := mockWxApp.IsValidAccess(myAccessToken, myOpenID)

	t.Log(ok)
}
