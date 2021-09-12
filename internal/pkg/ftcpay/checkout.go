package ftcpay

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
)

type Counter struct {
	BaseAccount account.BaseAccount
	price.CheckoutItem
	PayMethod enum.PayMethod
	WxAppID   null.String
}

func (c Counter) buildOrder(k enum.OrderKind) (subs.Order, error) {
	orderID, err := ids.OrderID()
	if err != nil {
		return subs.Order{}, err
	}

	return subs.Order{
		ID:            orderID,
		UserIDs:       c.BaseAccount.CompoundIDs(),
		PlanID:        c.Price.ID,
		DiscountID:    null.NewString(c.Offer.ID, c.Offer.ID != ""),
		Price:         c.Price.UnitAmount,
		Edition:       c.Price.Edition,
		Charge:        price.NewCharge(c.Price, c.Offer),
		Kind:          k,
		PaymentMethod: c.PayMethod,
		WxAppID:       c.WxAppID,
		DatePeriod:    dt.DatePeriod{},
		CreatedAt:     chrono.TimeNow(),
		ConfirmedAt:   chrono.Time{},
		LiveMode:      c.Price.LiveMode,
	}, nil
}

func (c Counter) PaymentIntent(m reader.Membership) (subs.PaymentIntent, error) {

	if !m.EnjoyOffer(c.Offer) {
		return subs.PaymentIntent{}, errors.New("discount offer selected is not applicable to current membership")
	}

	orderKind, err := m.OrderKindOfOneTime(c.Price.Edition)
	if err != nil {
		return subs.PaymentIntent{}, err
	}

	order, err := c.buildOrder(orderKind)
	if err != nil {
		return subs.PaymentIntent{}, err
	}

	return subs.PaymentIntent{
		Pricing:    c.Price,
		Offer:      c.Offer,
		Order:      order,
		Membership: m,
	}, nil
}
