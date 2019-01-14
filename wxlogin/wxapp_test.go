package wxlogin

import (
	"os"
	"testing"
)

const (
	myUUID         = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	myUnionID      = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	myOpenID       = "ob7fA0h69OO0sTLyQQpYc55iF_P0"
	myWxCode       = "001hPJNE0xvMjl25EaLE0k4PNE0hPJNl"
	myAccessToken  = "***REMOVED***"
	myRefreshToken = "***REMOVED***"
	myNickname     = "Victor"
	myAvatar       = "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLB34sBwSiaL3GJmejqDUqJw4CZ8Qs0ztibsRu6wzMpg7jg5icxWKwxF73ssZUmXmee1MvSvaZ6iaqs1A/132"
)

var mockWxApp = WxApp{
	AppID:     os.Getenv("WXPAY_APPID"),
	AppSecret: os.Getenv("WXPAY_APPSECRET"),
}

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
