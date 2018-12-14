package wxlogin

import (
	"database/sql"
	"encoding/base64"
	"os"
	"testing"

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

var mockReqClient = util.RequestClient{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  "golang mocker",
}

var mockWxClient = Client{
	AppID:     os.Getenv("WXPAY_APPID"),
	AppSecret: os.Getenv("WXPAY_APPSECRET"),
}

const (
	mockCode   = "001hPJNE0xvMjl25EaLE0k4PNE0hPJNl"
	mockUserID = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
)

var mockAccess = OAuthAccess{
	AccessToken:  "***REMOVED***",
	ExpiresIn:    7200,
	RefreshToken: "***REMOVED***",
	OpenID:       "ob7fA0h69OO0sTLyQQpYc55iF_P0",
	Scope:        "snsapi_userinfo",
	UnionID:      null.StringFrom("ogfvwjk6bFqv2yQpOrac0J3PqA0o"),
}

var mockUserInfo = UserInfo{
	UnionID:    "ogfvwjk6bFqv2yQpOrac0J3PqA0o",
	OpenID:     "ob7fA0h69OO0sTLyQQpYc55iF_P0",
	NickName:   "Victor",
	HeadImgURL: "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLB34sBwSiaL3GJmejqDUqJw4CZ8Qs0ztibsRu6wzMpg7jg5icxWKwxF73ssZUmXmee1MvSvaZ6iaqs1A/132",
	Gender:     0,
	Country:    "",
	Province:   "",
	City:       "",
	Privileges: []string{},
}

func TestDecodeToken(t *testing.T) {
	b, err := base64.RawURLEncoding.DecodeString(mockAccess.AccessToken)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Token in hex: %x\n", b)
}
