package paywall

import (
	"github.com/FTChinese/go-rest"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
type Subscription struct {
	OrderID       string // At first I mean to unexport it, but it is a scan target so must be exported.
	CompoundID    string // Use FTCUserID if it is valid, then use UnionID if it is valid, then throw error. This column is acting only as non-null constraint.
	FTCUserID     null.String
	UnionID       null.String
	TierToBuy     enum.Tier
	BillingCycle  enum.Cycle // Calculate expiration date
	ListPrice     float64
	NetPrice      float64
	PaymentMethod enum.PayMethod
	CreatedAt     chrono.Time // When the order is created.
	ConfirmedAt   chrono.Time // When the payment is confirmed.
	IsRenewal     bool        // If this order is used to renew membership. Determined the moment notification is received. Mostly used for data analysis and email.
	StartDate     chrono.Date // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date // Membership end date for this order. Depends on start date.
}

// NewWxpaySubs creates a new Subscription with payment method set to Wechat.
// Note wechat login and wechat pay we talked here are two totally non-related things.
func NewWxpaySubs(ftcID null.String, unionID null.String, p Plan) (Subscription, error) {
	s := Subscription{
		FTCUserID:     ftcID,
		UnionID:       unionID,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
		PaymentMethod: enum.PayMethodWx,
	}

	compoundID, err := s.PickCompoundID()
	if err != nil {
		return s, err
	}

	s.CompoundID = compoundID

	if err := s.generateOrderID(); err != nil {
		return s, err
	}

	return s, nil
}

// NewAlipaySubs creates a new Subscription with payment method set to Alipay.
func NewAlipaySubs(ftcID null.String, unionID null.String, p Plan) (Subscription, error) {
	s := Subscription{
		FTCUserID:     ftcID,
		UnionID:       unionID,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
		PaymentMethod: enum.PayMethodAli,
	}

	if ftcID.Valid {
		s.CompoundID = ftcID.String
	} else if unionID.Valid {
		s.CompoundID = unionID.String
	} else {
		return s, errors.New("ftc user id and union id should not both be null")
	}

	compoundID, err := s.PickCompoundID()
	if err != nil {
		return s, err
	}

	s.CompoundID = compoundID

	err = s.generateOrderID()
	if err != nil {
		return s, err
	}

	return s, nil
}

func (s Subscription) PickCompoundID () (string, error) {
	if s.FTCUserID.Valid {
		return s.FTCUserID.String, nil
	} else if s.UnionID.Valid {
		return s.UnionID.String, nil
	} else {
		return "", errors.New("ftc user id and union id should not both be null")
	}
}
// GenerateOrderID creates an id for this order. The order id is created only created upon the initial call of this method. Multiple calls won't change the this order's id.
func (s *Subscription) generateOrderID() error {

	id, err := gorest.RandomHex(8)
	if err != nil {
		return err
	}

	s.OrderID = "FT" + strings.ToUpper(id)

	return nil
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

// ConfirmWithDuration populate a subscription's ConfirmedAt, StartDate, EndDate and IsRenewal based on a user's current membership duration.
// It picks whichever comes last from Duration.ExpireDate
// or confirmedAt.
func (s Subscription) ConfirmWithDuration(dur Duration, confirmedAt time.Time) (Subscription, error) {
	s.ConfirmedAt = chrono.TimeFrom(confirmedAt)

	dur.NormalizeDate()

	s.IsRenewal = dur.ExpireDate.After(confirmedAt)

	var startTime time.Time
	// If a membership's ExpireDate is after confirmedAt,
	// use expireDate as new subscription's startTime;
	// otherwise use the confirmedAt as startTime.
	// Since the zero value of ExpireDate is always before
	// confirmedAt, the `else` branch always wins.
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
