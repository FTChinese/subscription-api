package paywall

import (
	"strconv"
	"strings"
	"time"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
type Subscription struct {
	UserID        string // It might be FTC UUID or Wechat union id, depnding on the login method.
	OrderID       string
	LoginMethod   enum.LoginMethod // Determine login method.
	TierToBuy     enum.Tier
	BillingCycle  enum.Cycle // Caculate expiration date
	ListPrice     float64
	NetPrice      float64
	PaymentMethod enum.PayMethod
	CreatedAt     chrono.Time // When the order is created.
	ConfirmedAt   chrono.Time // When the payment is confirmed.
	IsRenewal     bool        // If this order is used to renew membership. Determined the moment notification is received. Mostly used for data anaylsis and email.
	StartDate     chrono.Date // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date // Membership end date for this order. Depends on start date.
}

// NewWxpaySubs creates a new Subscription with payment method set to Wechat.
// Note wechat login and wechat pay we talked here are two totally non-related things.
func NewWxpaySubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	s := Subscription{
		UserID:        userID,
		LoginMethod:   login,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
		PaymentMethod: enum.PayMethodWx,
	}

	s.GenerateOrderID()

	return s
}

// NewAlipaySubs creates a new Subscription with payment method set to Alipay.
func NewAlipaySubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	s := Subscription{
		UserID:        userID,
		LoginMethod:   login,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
		PaymentMethod: enum.PayMethodAli,
	}

	s.GenerateOrderID()

	return s
}

// GenerateOrderID creates an id for this order. The order id is created only created upon the initial call of this method. Multiple calls won't change the this order's id.
func (s *Subscription) GenerateOrderID() {
	if s.OrderID != "" {
		return
	}

	id, _ := gorest.RandomHex(8)

	s.OrderID = "FT" + strings.ToUpper(id)
}

// AliNetPrice converts Charged price to ailpay format
func (s Subscription) AliNetPrice() string {
	return strconv.FormatFloat(s.NetPrice, 'f', 2, 32)
}

// WxNetPrice converts Charged price to int64 in cent for comparison with wx notification.
func (s Subscription) WxNetPrice() int64 {
	return int64(s.NetPrice * 100)
}

// IsWxChargeMatched tests if the order's charge matches the one from wechat response.
func (s Subscription) IsWxChargeMatched(cent int64) bool {
	return s.WxNetPrice() == cent
}

// IsConfirmed checks if the order is confirmed.
func (s Subscription) IsConfirmed() bool {
	return !s.ConfirmedAt.IsZero()
}

// IsWxLogin Check if user logged in by Wechat account.
func (s Subscription) IsWxLogin() bool {
	return s.LoginMethod == enum.LoginMethodWx
}

// IsEmailLogin checks if user logged in by email.
func (s Subscription) IsEmailLogin() bool {
	return s.LoginMethod == enum.LoginMethodEmail
}

// GetUnionID creates a nullable union id from user id if user logged in via wechat.
func (s Subscription) GetUnionID() null.String {
	if s.IsWxLogin() {
		return null.StringFrom(s.UserID)
	}

	return null.String{}
}

// ConfirmWithDuration populate a subscripiton's ConfirmedAt, StartDate, EndDate and IsRenewal based on a user's current membership duration.
// Current membership might not exists, but the duration is still a valid value since the zero value can be treated as a non-existing membership.
func (s Subscription) ConfirmWithDuration(dur Duration, confirmedAt time.Time) (Subscription, error) {
	s.ConfirmedAt = chrono.TimeFrom(confirmedAt)

	dur.NormalizeDate()

	s.IsRenewal = dur.ExpireDate.After(confirmedAt)

	var startTime time.Time
	if s.IsRenewal {
		startTime = dur.ExpireDate.Time
	} else {
		startTime = confirmedAt
	}

	expireTime, err := s.BillingCycle.TimeAfterACycle(startTime)
	if err != nil {
		return s, err
	}

	s.StartDate = chrono.DateFrom(startTime)
	s.EndDate = chrono.DateFrom(expireTime)

	return s, nil
}
