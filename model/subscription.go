package model

import (
	"database/sql"
	"strconv"
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
type Subscription struct {
	OrderID       string
	TierToBuy     MemberTier
	BillingCycle  BillingCycle
	Price         float64
	TotalAmount   float64
	PaymentMethod PaymentMethod
	Currency      string
	CreatedAt     string // When the order is created.
	ConfirmedAt   string // When the payment is confirmed.
	IsRenewal     bool   // If this order is used to renew membership
	StartDate     string // Membership start date for this order
	EndDate       string // Membership end date for this order
	UserID        string
}

// WxTotalFee converts TotalAmount to int64 in cent for comparison with wx notification.
func (s Subscription) WxTotalFee() int64 {
	return int64(s.TotalAmount * 100)
}

// AliTotalAmount converts TotalAmount to ailpay format
func (s Subscription) AliTotalAmount() string {
	return strconv.FormatFloat(s.TotalAmount, 'f', 2, 32)
}

// deduceExpireTime deduces membership expiration time based on when it is confirmed and the billing cycle.
// Add one day more to accomodate timezone change
func (s Subscription) deduceExpireTime(t time.Time) time.Time {
	switch s.BillingCycle {
	case Yearly:
		return t.AddDate(1, 0, 1)

	case Monthly:
		return t.AddDate(0, 1, 1)
	}

	return t
}

// CreatedAtCN turns creation time into Chinese text and format.
func (s Subscription) CreatedAtCN() string {
	return util.FormatShanghai.FromISO8601(s.CreatedAt)
}

// SaveSubscription saves a new subscription order.
// At this moment, you should already know if this subscription is
// a renewal of a new one, based on current Membership's expire_date.
func (env Env) SaveSubscription(s Subscription, c util.RequestClient) error {
	query := `
	INSERT INTO premium.ftc_trade
	SET trade_no = ?,
		trade_price = ?,
		trade_amount = ?,
		user_id = ?,
		tier_to_buy = ?,
		billing_cycle = ?,
		payment_method = ?,
		is_renewal = ?,
		created_utc = UTC_TIMESTAMP(),
		client_type = ?,
		client_version = ?,
		user_ip_bin = INET6_ATON(?),
		user_agent = NULLIF(?, '')`

	_, err := env.DB.Exec(query,
		s.OrderID,
		s.Price,
		s.TotalAmount,
		s.UserID,
		string(s.TierToBuy),
		string(s.BillingCycle),
		string(s.PaymentMethod),
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

// FindSubscription tries to find an order to verify the authenticity of a subscription order.
func (env Env) FindSubscription(orderID string) (Subscription, error) {
	query := `
	SELECT trade_no AS orderId,
		trade_price AS price,
		trade_amount AS totalAmount,
		user_id AS userId,
		IFNULL(tier_to_buy, '') AS tierToBuy,
		IFNULL(billing_cycle, '') AS billingCycle,
		IFNULL(payment_method, '') AS paymentMethod,
		is_renewal AS isRenewal,
		IFNULL(created_utc, '') AS createdAt,
		IFNULL(confirmed_utc, '') AS confirmedAt,
		IFNULL(start_date, '') AS startDate,
		IFNULL(end_date, '') AS endDate
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	var s Subscription
	err := env.DB.QueryRow(query, orderID).Scan(
		&s.OrderID,
		&s.Price,
		&s.TotalAmount,
		&s.UserID,
		&s.TierToBuy,
		&s.BillingCycle,
		&s.PaymentMethod,
		&s.IsRenewal,
		&s.CreatedAt,
		&s.ConfirmedAt,
	)

	if err != nil {
		logger.WithField("location", "Find subscription").Error(err)
		return s, err
	}

	if s.CreatedAt != "" {
		s.CreatedAt = util.ISO8601UTC.FromDatetime(s.CreatedAt, nil)
	}

	if s.ConfirmedAt != "" {
		s.ConfirmedAt = util.ISO8601UTC.FromDatetime(s.ConfirmedAt, nil)
	}

	return s, nil
}

// Confirm updates a subscription order's ConfirmedAt, StartDate
// and EndDate based on passed in confirmation time.
// Fortunately StartDate and EndDate uses YYYY-MM-DD format, which
// conforms to SQL DATE type. So we do not need to convert it.
func (s Subscription) confirm(t time.Time) Subscription {
	expireTime := s.deduceExpireTime(t)

	s.ConfirmedAt = util.ISO8601UTC.FromTime(t)
	s.StartDate = util.SQLDateUTC.FromTime(t)
	s.EndDate = util.SQLDateUTC.FromTime(expireTime)

	return s
}

// Renew extends a membership.
func (s Subscription) renew(member Membership) (Subscription, error) {
	willEnd, err := util.ParseSQLDate(member.Expire)

	if err != nil {
		return s, err
	}

	expireTime := s.deduceExpireTime(willEnd)

	s.StartDate = util.SQLDateUTC.FromTime(willEnd)
	s.EndDate = util.SQLDateUTC.FromTime(expireTime)

	return s, nil
}

// ConfirmSubscription marks an order as completed and create a member or renew membership.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
// NOTE: The passed in Subscription must be one retrieved from database. Otherwise you should never call this method.
func (env Env) ConfirmSubscription(s Subscription, confirmTime time.Time) (Subscription, error) {
	subs := s.confirm(confirmTime)

	if !s.IsRenewal {
		return subs, nil
	}

	member, err := env.FindMember(s.UserID)

	// If err is SqlNoRows error, do not use this subs.
	if err != nil {
		// If the error is sql.ErrNoRows, `subs` is valid.
		if err == sql.ErrNoRows {
			return subs, nil
		}
		return subs, err
	}

	// Membership already exists. This subscription is used for renewal.
	renewalSubs, err := subs.renew(member)

	if err != nil {
		return subs, err
	}

	return renewalSubs, nil
}

// CreateOrUpdateMember updates subscription order and create/update membership in one transaction.
func (env Env) CreateOrUpdateMember(subs Subscription) error {
	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("location", "CreateOrUpdateMember begin Transaction").Error(err)
		return err
	}

	stmtUpdate := `
	UPDATE premium.ftc_trade
	SET confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`

	// IMPORTANT: Do not forget to convert ISO8601 string to SQL DATETIME!
	// If you forgot to do so, MySQL won't given you error details.
	confirmed := util.SQLDatetimeUTC.FromISO8601(subs.ConfirmedAt)

	_, updateErr := tx.Exec(stmtUpdate,
		confirmed,
		subs.StartDate,
		subs.EndDate,
		subs.OrderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
		logger.WithField("location", "CreateOrUpdateMember update order").Error(err)
	}

	stmtCreate := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	_, createErr := tx.Exec(stmtCreate,
		subs.UserID,
		string(subs.TierToBuy),
		string(subs.BillingCycle),
		subs.StartDate,
		string(subs.TierToBuy),
		string(subs.BillingCycle),
		subs.EndDate,
	)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("location", "CreateOrUpdateMember create or update membership").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("location", "CreateOrUpdateMember commit transaction`").Error(err)
		return err
	}

	return nil
}

// SendConfirmationLetter sends an email to user that current
// subscription order is confirmed, based on the order detials.
func (env Env) SendConfirmationLetter(subs Subscription) error {
	// 1. Find this user's personal data
	user, err := env.FindUser(subs.UserID)

	if err != nil {
		return err
	}

	// 2. Compose email content
	parcel, err := ComposeEmail(user, subs)
	if err != nil {
		return err
	}

	err = env.PostOffice.SendLetter(parcel)

	return err
}
