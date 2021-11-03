package wxoauth

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"testing"
)

func TestClient_RefreshAccess(t *testing.T) {
	faker.MustSetupViper()

	client := NewClient(config.MustWxNativeApp())

	accResp, err := client.RefreshAccess("")

	if err != nil {
		t.Error(err)
		return
	}

	// {
	// AccessToken:***REMOVED***
	// ExpiresIn:7200
	// RefreshToken:50_L89wZSqDFdDSAf6THq0YaTr0HgJxWO65Ez0n5fFVKw4-hQtPRzDlzWFapSBrXB5JmfQewjt1rrvwRT31bdLgV_8QdJeH6kX_q47XeJ473Lc
	// OpenID:ob7fA0h69OO0sTLyQQpYc55iF_P0
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

	client := NewClient(config.MustWxNativeApp())

	isValid := client.IsAccessTokenValid(wxlogin.UserInfoParams{
		AccessToken: "",
		OpenID:      "",
	})

	t.Logf("Valid %t", isValid)
}
