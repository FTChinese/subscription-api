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
	UnionID       null.String // If UnionID is valid, then UserID must be equal to it.
	OrderID       string
	TierToBuy     enum.Tier
	BillingCycle  enum.Cycle
	Price         float64
	TotalAmount   float64
	PaymentMethod enum.PayMethod
	Currency      string
	CreatedAt     util.ISODateTime // When the order is created.
	ConfirmedAt   util.ISODateTime // When the payment is confirmed.
	IsRenewal     bool             // If this order is used to renew membership
	StartDate     string           // Membership start date for this order
	EndDate       string           // Membership end date for this order
}

func newSubs(userID, unionID string, p Plan, method enum.PayMethod) Subscription {
	subs := Subscription{
		OrderID:       p.OrderID(),
		TierToBuy:     p.Tier,
		BillingCycle:  p.Cycle,
		Price:         p.Price,
		TotalAmount:   p.Price,
		PaymentMethod: method,
	}

	if userID != "" {
		subs.UserID = userID
		return subs
	}

	subs.UserID = unionID
	subs.UnionID = null.StringFrom(unionID)
	return subs
}

// NewWxSubs creates a new Subscription with payment method set to Wechat.
func NewWxSubs(userID, unionID string, p Plan) Subscription {
	return newSubs(userID, unionID, p, enum.Wxpay)
}

// NewAliSubs creates a new Subscription with payment method set to Alipay.
func NewAliSubs(userID, unionID string, p Plan) Subscription {
	return newSubs(userID, unionID, p, enum.Alipay)
}

// Check if user logged in by Wechat account.
func (s Subscription) isWxLogin() bool {
	return s.UnionID.Valid
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
func (s Subscription) CreatedAtCN() string {
	dtStr := string(s.CreatedAt)
	cst, err := util.ToCST.FromISO8601(dtStr)

	// If conversion failed, use the original date time string.
	if err != nil {
		return dtStr
	}

	return cst
}

// Confirm updates a subscription order's ConfirmedAt, StartDate
// and EndDate based on passed in confirmation time.
// Fortunately StartDate and EndDate uses YYYY-MM-DD format, which
// conforms to SQL DATE type. So we do not need to convert it.
func (s Subscription) withConfirmation(t time.Time) (Subscription, error) {

	// Calculate expiration time by adding one cycle to the confirmation time.
	expireTime, err := s.BillingCycle.TimeAfterACycle(t)

	// If expiration time cannot be deduced.
	// (minght be caused by wrong billing cycle)
	if err != nil {
		return s, err
	}

	// Convert the confirmation Time instance to ISO8601 string.
	s.ConfirmedAt = util.ISODateTime(util.ToISO8601UTC.FromTime(t))

	// Use confirmation time's year-month-date part as this order's subscriped beginning date.
	s.StartDate = util.ToSQLDateUTC.FromTime(t)
	// Use the expiration time's year-month-date part as this order's purchased ending date.
	s.EndDate = util.ToSQLDateUTC.FromTime(expireTime)

	return s, nil
}

// update subscription's StartDate and EndDate based on
// previous membership's expiration date
// after the subscription is confirmed.
func (s Subscription) withMembership(member Membership) (Subscription, error) {
	expireTime, err := util.ParseDateTime(member.ExpireDate)

	if err != nil {
		return s, err
	}

	// Add a cycle to current membership's expiration time
	// expireTime := s.deduceExpireTime(willEnd)
	extendedTime, err := s.BillingCycle.TimeAfterACycle(expireTime)

	if err != nil {
		return s, err
	}

	s.StartDate = util.ToSQLDateUTC.FromTime(expireTime)
	s.EndDate = util.ToSQLDateUTC.FromTime(extendedTime)

	return s, nil
}

// PlaceOrder creates a new order for a user
// and remembers if this order is used to
// renew existing membership or simply
// create a new one.
func (env Env) PlaceOrder(subs Subscription, c util.ClientApp) error {
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

	err = env.saveSubscription(subs, c)

	if err != nil {
		return err
	}

	return nil
}

// saveSubscription saves a new subscription order.
// At this moment, you should already know if this subscription is
// a renewal of a new one, based on current Membership's expire_date.
func (env Env) saveSubscription(s Subscription, c util.ClientApp) error {
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
		user_agent = ?`

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
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		payment_method AS paymentMethod,
		is_renewal AS isRenewal,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt,
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

	_, updateErr := tx.Exec(stmtUpdate,
		subs.ConfirmedAt,
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
