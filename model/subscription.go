package model

import (
	"database/sql"
	"strconv"
	"time"

	"gitlab.com/ftchinese/subscription-api/member"

	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
type Subscription struct {
	OrderID       string
	TierToBuy     member.Tier
	BillingCycle  member.Cycle
	Price         float64
	TotalAmount   float64
	PaymentMethod member.PayMethod
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

// CreatedAtCN turns creation time into Chinese text and format.
func (s Subscription) CreatedAtCN() string {
	return util.FormatShanghai.FromISO8601(s.CreatedAt)
}

// Confirm updates a subscription order's ConfirmedAt, StartDate
// and EndDate based on passed in confirmation time.
// Fortunately StartDate and EndDate uses YYYY-MM-DD format, which
// conforms to SQL DATE type. So we do not need to convert it.
func (s Subscription) withConfirmation(t time.Time) (Subscription, error) {
	// expireTime := s.deduceExpireTime(t)
	expireTime, err := s.BillingCycle.TimeAfterACycle(t)

	if err != nil {
		return s, err
	}
	s.ConfirmedAt = util.ISO8601UTC.FromTime(t)
	s.StartDate = util.SQLDateUTC.FromTime(t)
	s.EndDate = util.SQLDateUTC.FromTime(expireTime)

	return s, nil
}

// update subscription's StartDate and EndDate based on
// previous membership's expiration date
// after the subscription is confirmed.
func (s Subscription) withMembership(member Membership) (Subscription, error) {
	expireTime, err := util.ParseSQLDate(member.Expire)

	if err != nil {
		return s, err
	}

	// Add a cycle to current membership's expiration time
	// expireTime := s.deduceExpireTime(willEnd)
	extendedTime, err := s.BillingCycle.TimeAfterACycle(expireTime)

	if err != nil {
		return s, err
	}

	s.StartDate = util.SQLDateUTC.FromTime(expireTime)
	s.EndDate = util.SQLDateUTC.FromTime(extendedTime)

	return s, nil
}

// PlaceOrder creates a new order for a user
// and remembers if this order is used to
// renew existing membership or simply
// create a new one.
func (env Env) PlaceOrder(subs Subscription, c util.RequestClient) error {
	// Check if we could find the membership for this user.
	member, err := env.FindMember(subs.UserID)

	// If the membership if found.
	if err == nil {
		// If membership is not allowed to renew yet
		if !member.CanRenew(subs.BillingCycle) {
			return util.ErrRenewalForbidden
		}

		// If current membership is allowed to renew,
		// and membership is not expired yet,
		// we remember that this order is used for renewal.
		if !member.IsExpired() {
			subs.IsRenewal = true
		}
	}

	err = env.SaveSubscription(subs, c)

	if err != nil {
		return err
	}

	return nil
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
		&s.StartDate,
		&s.EndDate,
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

// ConfirmSubscription retrieves a previously saved subscription order
// and update its ConfirmedAt, StartDate, EndDate fields base on whether this order
// is used for new member or renewal of existing one.
// This step does not persist the updated subscription order since that operation
// needs to be done together with membeship
// persistent.
func (env Env) ConfirmSubscription(s Subscription, confirmTime time.Time) (Subscription, error) {
	subs, err := s.withConfirmation(confirmTime)

	if err != nil {
		return s, err
	}

	// If this is a new member, the subscrition information is complete.
	if !s.IsRenewal {
		return subs, nil
	}

	// If this is a renewal, we need to find the current membership's expiration date.
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
	subs, err = subs.withMembership(member)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

// CreateOrUpdateMember updates subscription order and create/update membership in one transaction.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
// NOTE: The passed in Subscription must be one retrieved from database. Otherwise you should never call this method.
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
		subs.TierToBuy,
		subs.BillingCycle,
		subs.StartDate,
		subs.TierToBuy,
		subs.BillingCycle,
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
