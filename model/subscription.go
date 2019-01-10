package model

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/guregu/null"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Subscription contains the details of a user's action to place an order.
// This is the centrum of the whole subscription process.
//
// The worfolow is as follows:
//
// 1. Client send a request to create an order;
// 2. API will get the user id from request header; therefore user id must be incluced in request header.
// 3. Use the user id to query if a membership exists in database.
// 4. If the membership exists, check the membership's expiration date to see whether this subscription order is used to renew membership duration. The IsRenwal field will be true is user is trying to renew hist current membership; otherwise it will be false (membership does not exist, membership already expired, etc.)
// The subscription order creation process ends here. The PlaceOrder method incorporates those process in one place.
//
// The next step is confirmation process:
//
// 1. The payment provider notifies our server that an order is confirmed.
// 2. Retrieve previously saved subscription order based the order id extracted from notification message.
// 3. Update this subscription order's confirmation time, take the confirmation time as this order's start date and deduce the expiration date based on the confirmation time.
// 4. If the subscription order is used to renew a membership (remember we have a IsRenewal field in the order creation process?),
// go on to update the order's start date as membership expiration date and deduce end date based on this start date.
// 5. After all field is updated, we begin to persist the data into database, using SQL's transacation so that subscription order's confirmation data and a user's membership data are saved in one shot, or fail together.
type Subscription struct {
	UserID        string
	LoginMethod   enum.LoginMethod
	OrderID       string
	TierToBuy     enum.Tier
	BillingCycle  enum.Cycle
	Price         float64
	TotalAmount   float64
	PaymentMethod enum.PayMethod
	Currency      string
	CreatedAt     util.Time // When the order is created.
	ConfirmedAt   util.Time // When the payment is confirmed.
	IsRenewal     bool      // If this order is used to renew membership
	StartDate     util.Date // Membership start date for this order
	EndDate       util.Date // Membership end date for this order
}

// NewWxSubs creates a new Subscription with payment method set to Wechat.
// Note wechat login and wechat pay we talked here are two totally non-related things.
func NewWxSubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	return Subscription{
		UserID:        userID,
		LoginMethod:   login,
		OrderID:       p.OrderID(),
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Wxpay,
	}
}

// NewAliSubs creates a new Subscription with payment method set to Alipay.
func NewAliSubs(userID string, p Plan, login enum.LoginMethod) Subscription {
	return Subscription{
		UserID:        userID,
		LoginMethod:   login,
		OrderID:       p.OrderID(),
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: enum.Alipay,
	}
}

// Check if user logged in by Wechat account.
func (s Subscription) isWxLogin() bool {
	return s.LoginMethod == enum.WechatLogin
}

// WxTotalFee converts TotalAmount to int64 in cent for comparison with wx notification.
func (s Subscription) WxTotalFee() int64 {
	return int64(s.TotalAmount * 100)
}

// AliTotalAmount converts TotalAmount to ailpay format
func (s Subscription) AliTotalAmount() string {
	return strconv.FormatFloat(s.TotalAmount, 'f', 2, 32)
}

// CreatedAtCN turns creation time into China Stadnard Time in Chinese text.
// func (s Subscription) CreatedAtCN() string {
// 	dtStr := string(s.CreatedAt)
// 	cst, err := util.ToCST.FromISO8601(dtStr)

// 	// If conversion failed, use the original date time string.
// 	if err != nil {
// 		return dtStr
// 	}

// 	return cst
// }

// Confirm updates a subscription order's ConfirmedAt, StartDate
// and EndDate based on passed in confirmation time.
// Fortunately StartDate and EndDate uses YYYY-MM-DD format, which
// conforms to SQL DATE type. So we do not need to convert it.
func (s Subscription) withConfirmation(t time.Time) (Subscription, error) {

	s.ConfirmedAt = util.TimeFrom(t)
	s.StartDate = util.DateFrom(t)

	// Calculate expiration time by adding one cycle to the confirmation time.
	expireTime, err := s.BillingCycle.EndingTime(t)

	// If expiration time cannot be deduced.
	// (minght be caused by wrong billing cycle)
	if err != nil {
		return s, err
	}

	// Use the expiration time's year-month-date part as this order's purchased ending date.
	s.EndDate = util.DateFrom(expireTime)

	return s, nil
}

// update subscription's StartDate and EndDate based on
// previous membership's expiration date
// after the subscription is confirmed.
// Can this method only if previous membership
// is not expired yet.
func (s Subscription) withMembership(member Membership) (Subscription, error) {
	s.IsRenewal = !member.isExpired()

	s.StartDate = member.ExpireDate

	// Add a cycle to current membership's expiration time
	// expireTime := s.deduceExpireTime(willEnd)
	expireTime, err := s.BillingCycle.EndingTime(member.ExpireDate.Time)

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

	// Try to find out if this subscription order is created by an existing member.
	member, err := env.findMember(s)

	// If err is SqlNoRows error, do not use this subs.
	if err != nil {
		// If the error is sql.ErrNoRows, `subs` is valid.
		logger.WithField("trace", "ConfirmSubscription").Error(err)

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

// CreateMembership updates subscription order and create/update membership in one transaction.
// Confirm order and create/renew a new member should be an all-or-nothing operation.
// Or update membership duration.
// NOTE: The passed in Subscription must be one retrieved from database. Otherwise you should never call this method.
func (env Env) CreateMembership(subs Subscription) error {
	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("location", "CreateOrUpdateMember begin Transaction").Error(err)
		return err
	}

	// Update subscription order.
	stmtUpdate := `
	UPDATE premium.ftc_trade
	SET is_renewal = ?,
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`

	_, updateErr := tx.Exec(stmtUpdate,
		subs.IsRenewal,
		subs.ConfirmedAt,
		subs.StartDate,
		subs.EndDate,
		subs.OrderID,
	)

	if updateErr != nil {
		_ = tx.Rollback()
		logger.WithField("location", "CreateOrUpdateMember update order").Error(err)
	}

	// Create or extend membership.
	stmtCreate := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?
	ON DUPLICATE KEY UPDATE
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	var unionID null.String
	if subs.isWxLogin() {
		unionID = null.StringFrom(subs.UserID)
	}

	_, createErr := tx.Exec(stmtCreate,
		subs.UserID,
		unionID,
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
		subs.TierToBuy,
		subs.BillingCycle,
		subs.EndDate,
	)

	if createErr != nil {
		_ = tx.Rollback()

		logger.WithField("trace", "CreateOrUpdateMember create or update membership").Error(err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "CreateOrUpdateMember commit transaction`").Error(err)
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
	parcel, err := user.ComposeParcel(subs)
	if err != nil {
		return err
	}

	err = env.PostMan.Deliver(parcel)

	return err
}
