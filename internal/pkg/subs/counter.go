package subs

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
)

// Counter collects all information to create a one-time purchase.
type Counter struct {
	BaseAccount account.BaseAccount
	pw.CartItemFtc
	PayMethod enum.PayMethod
	WxAppID   null.String
}

func (c Counter) buildOrder(k enum.OrderKind) (Order, error) {
	orderID, err := ids.OrderID()
	if err != nil {
		return Order{}, err
	}

	ymd := c.PeriodCount()

	return Order{
		ID:            orderID,
		UserIDs:       c.BaseAccount.CompoundIDs(),
		Tier:          c.Price.Tier,
		Cycle:         c.Price.Cycle,
		Kind:          k,
		OriginalPrice: c.Price.UnitAmount,
		PayableAmount: c.PayableAmount(),
		PaymentMethod: c.PayMethod,
		YearsCount:    ymd.Years,
		MonthsCount:   ymd.Months,
		DaysCount:     ymd.Months,
		WxAppID:       c.WxAppID,
		ConfirmedAt:   chrono.Time{},
		CreatedAt:     chrono.TimeNow(),
		StartDate:     chrono.Date{},
		EndDate:       chrono.Date{},
	}, nil
}

func (c Counter) PaymentIntent(m reader.Membership) (PaymentIntent, error) {

	if !m.EnjoyOffer(c.Offer) {
		return PaymentIntent{}, errors.New("discount offer selected is not applicable to current membership")
	}

	orderKind, err := m.OrderKindOfOneTime(c.Price.Edition)
	if err != nil {
		return PaymentIntent{}, err
	}

	order, err := c.buildOrder(orderKind)
	if err != nil {
		return PaymentIntent{}, err
	}

	return PaymentIntent{
		Price:      c.Price,
		Offer:      c.Offer,
		Order:      order,
		Membership: m,
	}, nil
}
