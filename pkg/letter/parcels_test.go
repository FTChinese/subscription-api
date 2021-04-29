package letter

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/pkg/config"
	"testing"
)

func TestDeliverParcel(t *testing.T) {
	config.MustSetupViper()

	conn := config.MustGetHanqiConn()
	t.Logf("%v", conn)

	postman := postoffice.New(conn)

	parcel, err := VerificationParcel(CtxVerification{
		UserName: "Victor",
		Email:    "neefrankie@163.com",
		Link:     "https://user.ftchinese.com",
		IsSignUp: true,
	})

	if err != nil {
		t.Error(err)
		return
	}

	err = postman.Deliver(parcel)
	if err != nil {
		t.Error(err)
	}
}
