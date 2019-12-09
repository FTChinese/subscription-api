package subscription

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"testing"
	"time"
)

func TestConfirmationBuilder_Build(t *testing.T) {
	ftcID := uuid.New().String()
	//unionID, _ := gorest.RandomBase64(21)

	oBuilder := NewOrderBuilder(reader.MemberID{
		CompoundID: ftcID,
		FtcID:      null.StringFrom(ftcID),
	}).
		SetPlan(yearlyStandard).
		SetPayMethod(enum.PayMethodWx).
		SetMembership(Membership{}).
		SetWallet(Wallet{})

	if err := oBuilder.Build(); err != nil {
		t.Error(err)
	}

	order, _ := oBuilder.Order()

	cBuilder := NewConfirmationBuilder(PaymentResult{
		Amount:      order.AmountInCent(true),
		OrderID:     order.ID,
		ConfirmedAt: time.Now(),
	}, true).
		SetMembership(Membership{}).
		SetOrder(order)

	if err := cBuilder.Build(); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", cBuilder.ConfirmedOrder())
}
