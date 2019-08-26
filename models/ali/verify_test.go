package ali

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/smartwalle/alipay"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"testing"
	"time"
)

func TestGetPaymentResult(t *testing.T) {
	orderID, err := paywall.GenerateOrderID()
	if err != nil {
		t.Error(err)
	}

	n := &alipay.TradeNotification{
		TotalAmount: "258.00",
		OutTradeNo:  orderID,
		GmtPayment:  time.Now().In(time.UTC).Format(chrono.SQLDateTime),
	}

	payResult, err := GetPaymentResult(n)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", payResult)

	order := paywall.Order{
		Amount: 258.00,
	}

	if order.AmountInCent() == payResult.Amount {
		t.Log("Equal")
	} else {
		t.Error("not equal")
	}
}
