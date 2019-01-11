package wepay

import (
	"strings"
	"testing"

	"github.com/guregu/null"
	"github.com/icrowley/fake"
)

func TestNewNotification(t *testing.T) {
	resp := createNotificationResponse()

	p, err := ParseResponse(mockClient, strings.NewReader(resp))
	if err != nil {
		t.Error(err)
		return
	}

	noti := NewNotification(p)

	t.Logf("Notification; %+v\n", noti)
}
func TestSaveNotification(t *testing.T) {
	n := Notification{
		StatusCode:    "SUCCESS",
		StatusMessage: "OK",
		AppID:         null.StringFrom("wx8888888888888888"),
		MID:           null.StringFrom("1900000109"),
		Nonce:         null.StringFrom(fake.CharactersN(32)),
		Signature:     null.StringFrom(mockSignature),
		IsSuccess:     true,
		ResultCode:    null.StringFrom("SUCCESS"),
		OpenID:        null.StringFrom(fake.CharactersN(32)),
		IsSubscribed:  false,
		TradeType:     null.StringFrom("APP"),
		BankType:      null.StringFrom("CMC"),
		TotalFee:      null.IntFrom(25800),
		Currency:      null.StringFrom("CNY"),
		TransactionID: null.StringFrom(fake.CharactersN(32)),
		FTCOrderID:    null.StringFrom(fake.CharactersN(32)),
		TimeEnd:       null.StringFrom("20091225091010"),
	}

	err := devEnv.SaveNotification(n)
	if err != nil {
		t.Error(err)
	}
}
