package wxlogin

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	uuid "github.com/satori/go.uuid"

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
	myUUID         = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	myUnionID      = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	myOpenID       = "ob7fA0h69OO0sTLyQQpYc55iF_P0"
	myWxCode       = "001hPJNE0xvMjl25EaLE0k4PNE0hPJNl"
	myAccessToken  = "16_Ix0E3WfWs9u5Rh9f-lB7_LgsQJ4zm1eodolFJpSzoQibTAuhIlp682vDmkZSaYIjD9gekOa1zQl-6c6S_CrN_cN9vx9mybwXNVgFbwPMMwM"
	myRefreshToken = "16_IlmA9eLGjJw7gBKBT48wff1V1hAYAdpmIqUAypspepm6DsQ6kkcLeZmP932s9PcKp1WM5P_1YwUNQqF-29B_0CqGTqMpWkaaiNSYp26MmB4"
	myNickname     = "Victor"
	myAvatar       = "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLB34sBwSiaL3GJmejqDUqJw4CZ8Qs0ztibsRu6wzMpg7jg5icxWKwxF73ssZUmXmee1MvSvaZ6iaqs1A/132"
)

var mockClient = util.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.2",
	UserIP:     fake.IPv4(),
	UserAgent:  fake.UserAgent(),
}

var mockWxApp = WxApp{
	AppID:     os.Getenv("WXPAY_APPID"),
	AppSecret: os.Getenv("WXPAY_APPSECRET"),
}

var mockUUID = uuid.Must(uuid.NewV4()).String()

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

func newFTCMember() Membership {
	return Membership{
		UserID:     mockUUID,
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
		ExpireDate: util.DateNow(),
	}
}

func newWxMember() Membership {
	unionID := generateWxID()
	return Membership{
		UserID:     unionID,
		UnionID:    null.StringFrom(unionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
		ExpireDate: util.DateNow(),
	}
}

func newBoundMember() Membership {
	return Membership{
		UserID:     mockUUID,
		UnionID:    null.StringFrom(generateWxID()),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
		ExpireDate: util.DateNow(),
	}
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
	acc.createdAt = util.TimeNow()
	acc.updatedAt = util.TimeNow()
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

func TestDecodeToken(t *testing.T) {
	b, err := base64.RawURLEncoding.DecodeString(myAccessToken)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Token in hex: %x\n", b)
}
