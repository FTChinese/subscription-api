package wxoauth

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_RefreshAccess(t *testing.T) {
	faker.MustSetupViper()

	client := NewClient(config.MustWxNativeApp(), zaptest.NewLogger(t))

	accResp, err := client.RefreshAccess("")

	if err != nil {
		t.Error(err)
		return
	}

	// {
	// AccessToken:
	// ExpiresIn:7200
	// RefreshToken:
	// OpenID:
	//Scope:snsapi_userinfo
	// UnionID:{NullString:{String: Valid:false}}
	// RespStatus:{
	//		Code:0
	//		Message:
	//  }
	// }
	t.Logf("%+v", accResp)
}

func TestClient_IsAccessTokenValid(t *testing.T) {
	faker.MustSetupViper()

	client := NewClient(config.MustWxNativeApp(), zaptest.NewLogger(t))

	isValid := client.IsAccessTokenValid(wxlogin.UserInfoParams{
		AccessToken: "",
		OpenID:      "",
	})

	t.Logf("Valid %t", isValid)
}
