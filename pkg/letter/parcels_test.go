package letter

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"testing"
)

func TestDeliverParcel(t *testing.T) {
	faker.MustSetupViper()

	conn := config.MustGetHanqiConn()
	t.Logf("%v", conn)

	pm := postman.New(conn)

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

	err = pm.Deliver(parcel)
	if err != nil {
		t.Error(err)
	}
}
