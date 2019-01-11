package model

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/objcoding/wxpay"

	"github.com/guregu/null"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wepay"
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
	return Subscription{
		UserID:        userID,
		OrderID:       p.OrderID(),
		LoginMethod:   login,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Wxpay,
	}
}

// NewAlipaySubs creates a new Subscription with payment method set to Alipay.
func NewAlipaySubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	return Subscription{
		UserID:        userID,
		OrderID:       p.OrderID(),
		LoginMethod:   login,
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Alipay,
	}
}

// AliTotalAmount converts TotalAmount to ailpay format
func (s Subscription) AliTotalAmount() string {
	return strconv.FormatFloat(s.TotalAmount, 'f', 2, 32)
}

// WxTotalFee converts TotalAmount to int64 in cent for comparison with wx notification.
func (s Subscription) WxTotalFee() int64 {
	return int64(s.TotalAmount * 100)
}

// PrepayOrder creates a PrepayOrder from a subscription order for Wechat.
func (s Subscription) PrepayOrder(client *wxpay.Client, resp wxpay.Params) wepay.PrepayOrder {
	appID := resp.GetString("appid")
	partnerID := resp.GetString("mch_id")
	prepayID := resp.GetString("prepay_id")
	nonce, _ := util.RandomHex(10)
	pkg := "Sign=WXPay"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	p := make(wxpay.Params)
	p["appid"] = appID
	p["partnerid"] = partnerID
	p["prepayid"] = prepayID
	p["package"] = pkg
	p["noncestr"] = nonce
	p["timestamp"] = timestamp

	h := client.Sign(p)

	return wepay.PrepayOrder{
		FtcOrderID: s.OrderID,
		Price:      s.Price,
		AppID:      appID,
		PartnerID:  partnerID,
		PrepayID:   prepayID,
		Package:    pkg,
		Nonce:      nonce,
		Timestamp:  timestamp,
		Signature:  h,
	}
}

// Check if user logged in by Wechat account.
func (s Subscription) isWxLogin() bool {
	return s.LoginMethod == enum.WechatLogin
}

func (s Subscription) isEmailLogin() bool {
	return s.LoginMethod == enum.EmailLogin
}

func (s Subscription) getUnionID() null.String {
	if s.isWxLogin() {
		return null.StringFrom(s.UserID)
	}

	return null.String{}
}

// Build SQL query of membership depending on the login method; otherwise you cannot be sure the WHERE clause.
func (s Subscription) stmtMemberDuration() string {
	var whereCol string
	if s.isWxLogin() {
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

func (s Subscription) stmtMember() string {
	var whereCol string

	if s.isWxLogin() {
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

// withStartTime builds a subscription order's StartDate and EndDate based on the passed in starting time.
func (s Subscription) withStartTime(t time.Time) (Subscription, error) {
	s.StartDate = util.DateFrom(t)
	expireTime, err := s.BillingCycle.EndingTime(t)

	if err != nil {
		return s, err
	}

	s.EndDate = util.DateFrom(expireTime)

	return s, nil
}

// IsSubsAllowed checks if this user is allowed to purchase a subscritpion.
// If a user is a valid member, and the membership is not expired, and not within the allowed renewal period, deny the request.
func (env Env) IsSubsAllowed(subs Subscription) (bool, error) {
	member, err := env.findMember(subs)

	if err != nil {
		// If this user is not a member yet.
		if err == sql.ErrNoRows {
			return true, nil
		}

		logger.WithField("trace", "IsSubsAllowed").Error(err)
		// If any other unkonw error occurred
		return false, err
	}

	// Do not allow a subscribed user to change tiers.
	if subs.TierToBuy != member.Tier {
		logger.WithField("trace", "IsSubsAllowed").Error("Changing subscription tier is not supported.")
		return false, nil
	}

	// This user is/was a member.
	return member.canRenew(subs.BillingCycle), nil
}

// SaveSubscription saves a new subscription order.
// At this moment, you should already know if this subscription is
// a renewal of a new one, based on current Membership's expire_date.
func (env Env) SaveSubscription(s Subscription, c util.ClientApp) error {
	query := `
	INSERT INTO premium.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		user_id = ?,
		login_method = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		is_renewal = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = ?`

	_, err := env.DB.Exec(query,
		s.OrderID,
		s.Price,
		s.TotalAmount,
		s.UserID,
		s.LoginMethod,
		s.TierToBuy,
		s.BillingCycle,
		s.PaymentMethod,
		s.IsRenewal,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent,
	)

	if err != nil {
		logger.WithField("location", "New subscription").Error(err)
		return err
	}

	return nil
}

// VerifyWxNotification checks if price match, if already confirmed.
func (env Env) VerifyWxNotification(p wxpay.Params) error {
	orderID := p.GetString("out_trade_no")
	totalFee := p.GetInt64("total_fee")

	query := `
	SELECT trade_amount AS totalAmount
		confirmed_utc AS confirmedAt
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var amount float64
	var confirmedAt util.Time
	err := env.DB.QueryRow(query, orderID).Scan(
		&amount,
		&confirmedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	if !confirmedAt.IsZero() {
		logger.WithField("trace", "VerifyWxNotification").Error(ErrAlreadyConfirmed)

		return ErrAlreadyConfirmed
	}

	price := int64(amount * 100)

	if price != totalFee {
		logger.WithField("trace", "VerifyWxNotification").Infof("Paid price does not match. Should be %d, actual %d", price, totalFee)

		return ErrPriceMismatch
	}

	return nil
}

// FindSubscription tries to find an order to verify the authenticity of a subscription order.
func (env Env) FindSubscription(orderID string) (Subscription, error) {
	query := `
	SELECT trade_no AS orderId,
		trade_price AS price,
		trade_amount AS totalAmount,
		user_id AS userId,
		login_method AS loginMethod,
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		payment_method AS paymentMethod,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt,
		start_date AS startDate,
		end_date AS endDate
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var s Subscription
	err := env.DB.QueryRow(query, orderID).Scan(
		&s.OrderID,
		&s.Price,
		&s.TotalAmount,
		&s.UserID,
		&s.LoginMethod,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.PaymentMethod,
		&s.CreatedAt,
		&s.ConfirmedAt,
		&s.StartDate,
		&s.EndDate,
	)

	if err != nil {
		logger.WithField("trace", "FindSubscription").Error(err)
		return s, err
	}

	return s, nil
}

// ConfirmPayment handles payment notification with database locking.
func (env Env) ConfirmPayment(orderID string, confirmedAt time.Time) (Subscription, error) {

	var startTime time.Time

	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return Subscription{}, err
	}

	var subs Subscription
	errSubs := env.DB.QueryRow(stmtSubsLock, orderID).Scan(
		&subs.UserID,
		&subs.OrderID,
		&subs.LoginMethod,
		&subs.TierToBuy,
		&subs.BillingCycle,
		&subs.Price,
		&subs.TotalAmount,
		&subs.CreatedAt,
		&subs.ConfirmedAt,
	)

	if errSubs != nil {
		_ = tx.Rollback()
		if errSubs == sql.ErrNoRows {
			return subs, ErrOrderNotFound
		}
		return subs, errSubs
	}

	// Already confirmed.
	if !subs.ConfirmedAt.IsZero() {
		logger.WithField("trace", "ConfirmPayment").Infof("Order %s is already confirmed", orderID)

		_ = tx.Rollback()
		return subs, ErrAlreadyConfirmed
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Found order: %+v", subs)

	// Add confirmation time.
	subs.ConfirmedAt = util.TimeFrom(confirmedAt)

	// Start query membership expiration time.
	queryDuration := subs.stmtMemberDuration()

	var dur Duration
	errDur := env.DB.QueryRow(queryDuration, orderID).Scan(
		&dur.timestamp,
		&dur.ExpireDate,
	)

	if errDur != nil {
		// If no current membership is found for this order, confirmation time is the membership's start time.
		if errDur == sql.ErrNoRows {
			logger.WithField("trace", "ConfirmPayment").Infof("Member duration for user %s is not found", subs.UserID)

			subs.IsRenewal = false
			startTime = confirmedAt
		} else {
			_ = tx.Rollback()
			return subs, err
		}
	}

	dur.normalizeDate()
	// If membership is found, test if it is expired.
	if dur.isExpired() {
		subs.IsRenewal = false
		startTime = confirmedAt
	} else {
		subs.IsRenewal = true
		startTime = dur.ExpireDate.Time
	}

	subs, err = subs.withStartTime(startTime)
	if err != nil {
		return subs, err
	}

	logger.WithField("trace", "ConfirmPayment").Infof("Updated order: %+v", subs)

	// Update subscription order.
	_, updateErr := tx.Exec(stmtUpdateSubs,
		subs.IsRenewal,
		subs.ConfirmedAt,
		subs.StartDate,
		subs.EndDate,
		orderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "ConfirmPayment").Error(err)
	}

	// Create or extend membership.
	_, createErr := tx.Exec(stmtCreateMember,
		subs.UserID,
		subs.getUnionID(),
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
	)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "ConfirmPayment").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "ConfirmPayment").Error(err)
		return subs, err
	}

	return subs, nil
}
