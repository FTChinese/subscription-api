package wxlogin

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"

	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{DB: db}
}

var devEnv = newDevEnv()

const (
	mockUUID       = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	mockEmail      = "weiguo.ni@ftchinese.com"
	myUnionID      = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	myOpenID       = "ob7fA0h69OO0sTLyQQpYc55iF_P0"
	myWxCode       = "001hPJNE0xvMjl25EaLE0k4PNE0hPJNl"
	myAccessToken  = "***REMOVED***"
	myRefreshToken = "***REMOVED***"
)

var mockClient = util.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    null.StringFrom("1.1.2"),
	UserIP:     null.StringFrom(fake.IPv4()),
	UserAgent:  null.StringFrom(fake.UserAgent()),
}

var mockWxApp = WxApp{
	AppID:     os.Getenv("WXPAY_APPID"),
	AppSecret: os.Getenv("WXPAY_APPSECRET"),
}

func generateCode() string {
	code, _ := util.RandomBase64(24)
	return code
}

func generateToken() string {
	token, _ := util.RandomBase64(82)
	return token
}

func generateWxID() string {
	id, _ := util.RandomBase64(21)
	return id
}

func generateAvatarURL() string {
	return fmt.Sprintf("http://thirdwx.qlogo.cn/mmopen/vi_32/%s/132", fake.CharactersN(90))
}

type mockWxID struct {
	OpenID  string // Same OpenID under one app.
	UnionID string // Same user.
}

func newMockWxID() mockWxID {
	return mockWxID{
		OpenID:  generateWxID(),
		UnionID: generateToken(),
	}
}

func (wx mockWxID) generateAccess() OAuthAccess {
	acc := OAuthAccess{
		AccessToken:  generateToken(),
		ExpiresIn:    7200,
		RefreshToken: generateToken(),
		OpenID:       wx.OpenID,
		Scope:        "snsapi_userinfo",
		UnionID:      null.StringFrom(wx.UnionID),
	}
	acc.generateSessionID()
	acc.createdAt = time.Now()
	acc.updatedAt = time.Now()
	return acc
}

func (wx mockWxID) generateUserInfo() UserInfo {
	return UserInfo{
		UnionID:    wx.UnionID,
		NickName:   fake.UserName(),
		AvatarURL:  generateAvatarURL(),
		Sex:        randomdata.Number(0, 3),
		Country:    fake.Country(),
		Province:   fake.State(),
		City:       fake.City(),
		Privileges: []string{},
	}
}

func (wx mockWxID) generateMultiAccess() []OAuthAccess {
	var acc []OAuthAccess

	for i := 0; i < 3; i++ {
		a := wx.generateAccess()

		acc = append(acc, a)
	}

	return acc
}

func saveOneLogin(a OAuthAccess, u UserInfo) error {
	err := devEnv.SaveAccess(mockWxApp.AppID, a, mockClient)

	if err != nil {
		return err
	}

	err = devEnv.SaveUserInfo(u)

	if err != nil {
		return err
	}

	return nil
}

var nwgAccess = OAuthAccess{
	AccessToken:  myAccessToken,
	ExpiresIn:    7200,
	RefreshToken: myRefreshToken,
	OpenID:       myOpenID,
	Scope:        "snsapi_userinfo",
	UnionID:      null.StringFrom(myUnionID),
}

var nwgUserInfo = UserInfo{
	UnionID:    myUnionID,
	OpenID:     myOpenID,
	NickName:   "Victor",
	AvatarURL:  "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLB34sBwSiaL3GJmejqDUqJw4CZ8Qs0ztibsRu6wzMpg7jg5icxWKwxF73ssZUmXmee1MvSvaZ6iaqs1A/132",
	Sex:        0,
	Country:    "",
	Province:   "",
	City:       "",
	Privileges: []string{},
}

func TestDecodeToken(t *testing.T) {
	b, err := base64.RawURLEncoding.DecodeString(myAccessToken)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Token in hex: %x\n", b)
}
