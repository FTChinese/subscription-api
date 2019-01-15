package paywall

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/objcoding/wxpay"

	"github.com/smartwalle/alipay"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

const (
	aliCallbackURL = "http://www.ftacademy.cn/api/v1/callback/alipay"
	wxCallbackURL  = "http://www.ftacademy.cn/api/v1/callback/wxpay"
	aliProductCode = "QUICK_MSECURITY_PAY"
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
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
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
		ListPrice:     p.ListPrice,
		NetPrice:      p.NetPrice,
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

// AliNetPrice converts Charged price to ailpay format
func (s Subscription) AliNetPrice() string {
	return strconv.FormatFloat(s.NetPrice, 'f', 2, 32)
}

// AliAppPayParam builds parameters for ali app pay based on current subscription order.
func (s Subscription) AliAppPayParam(title string) alipay.AliPayParam {
	p := alipay.AliPayTradeAppPay{}
	p.NotifyURL = aliCallbackURL
	p.Subject = title
	p.OutTradeNo = s.OrderID
	p.TotalAmount = s.AliNetPrice()
	p.ProductCode = aliProductCode
	p.GoodsType = "0"

	return p
}

// AliAppPayResp builds the reponse for an app's request to pay by ali.
func (s Subscription) AliAppPayResp(param string) ali.AppPayResp {
	return ali.AppPayResp{
		FtcOrderID: s.OrderID,
		Price:      s.ListPrice,
		ListPrice:  s.ListPrice,
		NetPrice:   s.NetPrice,
		Param:      param,
	}
}

// WxNetPrice converts Charged price to int64 in cent for comparison with wx notification.
func (s Subscription) WxNetPrice() int64 {
	return int64(s.NetPrice * 100)
}

// WxUniOrderParam build the parameters to request for prepay id.
func (s Subscription) WxUniOrderParam(title, ip string) wxpay.Params {
	p := make(wxpay.Params)
	p.SetString("body", title)
	p.SetString("out_trade_no", s.OrderID)
	p.SetInt64("total_fee", s.WxNetPrice())
	p.SetString("spbill_create_ip", ip)
	p.SetString("notify_url", wxCallbackURL)
	p.SetString("trade_type", "APP")

	return p
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

// ConfirmWithDuration populate a subscripiton's ConfirmedAt, StartDate, EndDate and IsRenewal based on a user's current membership duration.
// Current membership might not exists, but the duration is still a valid value since the zero value can be treated as a non-existing membership.
func (s Subscription) ConfirmWithDuration(dur Duration, confirmedAt time.Time) (Subscription, error) {
	s.ConfirmedAt = util.TimeFrom(confirmedAt)

	dur.NormalizeDate()

	s.IsRenewal = dur.ExpireDate.After(confirmedAt)

	var startTime time.Time
	if s.IsRenewal {
		startTime = dur.ExpireDate.Time
	} else {
		startTime = confirmedAt
	}

	expireTime, err := s.BillingCycle.EndingTime(startTime)
	if err != nil {
		return s, err
	}

	s.StartDate = util.DateFrom(startTime)
	s.EndDate = util.DateFrom(expireTime)

	return s, nil
}
