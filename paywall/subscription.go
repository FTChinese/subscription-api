package paywall

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
type Subscription struct {
	UserID        string // Also used when creating member
	OrderID       string
	LoginMethod   enum.LoginMethod // Determine login method.
	TierToBuy     enum.Tier
	BillingCycle  enum.Cycle // Caculate expiration date
	Price         float64
	TotalAmount   float64
	PaymentMethod enum.PayMethod
	CreatedAt     util.Time // When the order is created.
	ConfirmedAt   util.Time // When the payment is confirmed.
	IsRenewal     bool      // If this order is used to renew membership. Determined the moment notification is received. Mostly used for data anaylsis and email.
	StartDate     util.Date // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       util.Date // Membership end date for this order. Depends on start date.
}

// NewWxpaySubs creates a new Subscription with payment method set to Wechat.
// Note wechat login and wechat pay we talked here are two totally non-related things.
func NewWxpaySubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	s := Subscription{
		UserID:        userID,
		LoginMethod:   login,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Wxpay,
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
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Alipay,
	}

	s.GenerateOrderID()

	return s
}

// GenerateOrderID creates an id for this order. The order id is created only created upon the initial call of this method. Multiple calls won't change the this order's id.
func (s *Subscription) GenerateOrderID() {
	if s.OrderID != "" {
		return
	}

	id, _ := util.RandomHex(8)

	s.OrderID = "FT" + strings.ToUpper(id)
}

// AliTotalAmount converts TotalAmount to ailpay format
func (s Subscription) AliTotalAmount() string {
	return strconv.FormatFloat(s.TotalAmount, 'f', 2, 32)
}

// WxTotalFee converts TotalAmount to int64 in cent for comparison with wx notification.
func (s Subscription) WxTotalFee() int64 {
	return int64(s.TotalAmount * 100)
}

// IsWxLogin Check if user logged in by Wechat account.
func (s Subscription) IsWxLogin() bool {
	return s.LoginMethod == enum.WechatLogin
}

// IsEmailLogin checks if user logged in by email.
func (s Subscription) IsEmailLogin() bool {
	return s.LoginMethod == enum.EmailLogin
}

// GetUnionID creates a nullable union id from user id if user logged in via wechat.
func (s Subscription) GetUnionID() null.String {
	if s.IsWxLogin() {
		return null.StringFrom(s.UserID)
	}

	return null.String{}
}

// StmtMemberDuration Build SQL query of membership depending on the login method; otherwise you cannot be sure the WHERE clause.
func (s Subscription) StmtMemberDuration() string {
	var whereCol string
	if s.IsWxLogin() {
		whereCol = "vip_id_alias"
	} else {
		whereCol = "vip_id"
	}

	return fmt.Sprintf(`
		SELECT expire_time AS expireTime,
			expire_date AS expireDate
		FROM premium.ftc_vip
		WHERE %s = ?
		LIMIT 1
		FOR UPDATE`, whereCol)
}

// StmtMember build SQL query of membership based on login method.
func (s Subscription) StmtMember() string {
	var whereCol string

	if s.IsWxLogin() {
		whereCol = "vip_id_alias"
	} else {
		whereCol = "vip_id"
	}

	return fmt.Sprintf(`
		SELECT vip_id AS userId,
			vip_id_alias AS unionId,
			vip_type AS vipType,
			member_tier AS memberTier,
			billing_cycle AS billingCyce,
			expire_time AS expireTime,
			expire_date AS expireDate
		FROM premium.ftc_vip
		WHERE %s = ?
		LIMIT 1`, whereCol)
}

// WithStartTime builds a subscription order's StartDate and EndDate based on the passed in starting time.
func (s Subscription) WithStartTime(t time.Time) (Subscription, error) {
	s.StartDate = util.DateFrom(t)
	expireTime, err := s.BillingCycle.EndingTime(t)

	if err != nil {
		return s, err
	}

	s.EndDate = util.DateFrom(expireTime)

	return s, nil
}
