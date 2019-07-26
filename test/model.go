package test

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"time"
)

type Model struct {
	db    *sql.DB
	query query.Builder
}

func NewModel() Model {
	return Model{
		db:    DB,
		query: query.NewBuilder(false),
	}
}

func (model Model) CreateGiftCard() paywall.GiftCard {
	c := GiftCard()

	q := `
	INSERT INTO premium.scratch_card
		SET serial_number = ?,
			auth_code = ?,
		    expire_time = UNIX_TIMESTAMP(?),
			tier = ?,
			cycle_unit = ?,
			cycle_value = ?`

	now := time.Now().AddDate(1, 0, 0)

	_, err := DB.Exec(q,
		GenCardSerial(),
		c.Code,
		now.Truncate(24*time.Hour),
		c.Tier,
		c.CycleUnit,
		c.CycleValue)

	if err != nil {
		panic(err)
	}

	return c
}

func (model Model) SaveSub(s paywall.Subscription) {

	var stmt = model.query.InsertSubs() + `,
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?`

	c := RandomClientApp()

	_, err := model.db.Exec(stmt,
		s.ID,
		s.User.CompoundID,
		s.User.FtcID,
		s.User.UnionID,
		s.ListPrice,
		s.Amount,
		s.Tier,
		s.Cycle,
		s.CycleCount,
		s.ExtraDays,
		s.Usage,
		s.PaymentMethod,
		s.WxAppID,
		c.ClientType,
		c.Version,
		c.UserIP,
		c.UserAgent,
		s.ConfirmedAt,
		s.StartDate,
		s.EndDate)

	if err != nil {
		panic(err)
	}
}

func (model Model) SaveMember(m paywall.Membership) {
	_, err := model.db.Exec(model.query.InsertMember(),
		m.ID,
		m.User.CompoundID,
		m.User.UnionID,
		m.TierCode(),
		m.ExpireDate.Unix(),
		m.User.FtcID,
		m.User.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		m.StripePlanID,
		m.AutoRenewal,
		m.Status)

	if err != nil {
		panic(err)
	}
}

func (model Model) UpdateMember(m paywall.Membership) {
	_, err := model.db.Exec(model.query.UpdateMember(),
		m.ID,
		m.TierCode(),
		m.ExpireDate.Unix(),
		m.Tier,
		m.Cycle,
		m.ExpireDate,
		m.PaymentMethod,
		m.StripeSubID,
		m.StripePlanID,
		m.AutoRenewal,
		m.Status,
		m.User.CompoundID,
		m.User.UnionID)

	if err != nil {
		panic(err)
	}
}

// Create a new order that is not confirmed.
func (model Model) CreateNewOrder(store *SubStore) paywall.Subscription {
	o := store.AddOrder(paywall.SubsKindCreate)
	model.SaveSub(o)

	return o
}

// CreateConfirmedOrder is used to test confirming an order in db.
func (model Model) CreateConfirmedOrder(store *SubStore) paywall.Subscription {
	o := model.CreateNewOrder(store)

	o, err := o.Confirm(store.Member, time.Now())
	if err != nil {
		panic(err)
	}

	return o
}

// CreateNewMember create a confirmed order and its
// membership but the membership is not save to db.
// This is used to test insert membership.
func (model Model) CreateNewMember(store *SubStore) paywall.Membership {
	o := store.AddOrder(paywall.SubsKindCreate)

	o, err := store.ConfirmOrder(o.ID)
	if err != nil {
		panic(err)
	}

	m, err := store.Member.FromAliOrWx(o)
	if err != nil {
		return m
	}
	store.Member = m

	// Save confirmed order
	model.SaveSub(o)

	return m
}

// CreateUpdateMember is used to test UpdateMember.
func (model Model) CreateUpdatedMember(store *SubStore) paywall.Membership {
	// First we create a new order and its membership.
	model.NewMemberCreated(store)

	// Create renewal order
	o := store.AddOrder(paywall.SubsKindRenew)

	// Confirm this order
	o, err := store.ConfirmOrder(o.ID)
	if err != nil {
		panic(err)
	}

	// Update membership.
	m, err := store.Member.FromAliOrWx(o)
	if err != nil {
		panic(err)
	}
	store.Member = m

	// Save the confirmed renewal order to db
	model.SaveSub(o)

	return m
}

// Create a new order and membership.
func (model Model) NewMemberCreated(store *SubStore) paywall.Subscription {
	// Create unconfirmed order
	o := store.AddOrder(paywall.SubsKindCreate)

	// Confirm this order
	o, err := store.ConfirmOrder(o.ID)
	if err != nil {
		panic(err)
	}

	m, err := store.Member.FromAliOrWx(o)
	if err != nil {
		panic(err)
	}
	store.Member = m

	// Save the confirmed order to db
	model.SaveSub(o)
	// Save the membership to db.
	model.SaveMember(store.Member)

	return o
}

// MemberRenewed renews a member and returns all the orders
// used to create and renew.
// Must call NewMemberCreate before calling this one.
func (model Model) MemberRenewed(store *SubStore) paywall.Subscription {

	if store.Member.IsZero() {
		panic("no membership exists to be renewed.")
	}

	// Then we create renewal order.
	o := store.AddOrder(paywall.SubsKindRenew)

	// Confirm this order.
	o, err := store.ConfirmOrder(o.ID)
	if err != nil {
		panic(err)
	}

	// Save the confirmed order
	model.SaveSub(o)
	// Update membership in db.
	model.UpdateMember(store.Member)

	// Return all the orders this user owns.
	return o
}

// MemberRenewedN renew a member for n times, and returned
// all the orders used perform all the renewal.
// The total number of orders returned is n + 1 since you n does not include the initial order used to create this membership.
func (model Model) MemberRenewedN(store *SubStore, n int) []paywall.Subscription {
	o := model.NewMemberCreated(store)

	orders := []paywall.Subscription{o}

	for i := 0; i < n; i++ {
		o := model.MemberRenewed(store)
		orders = append(orders, o)
	}

	return orders
}

func (model Model) UpgradeOrder(store *SubStore, n int) paywall.Subscription {
	o, err := store.UpgradeOrder(n)
	if err != nil {
		panic(err)
	}

	for _, v := range store.Orders {
		model.SaveSub(v)
	}

	model.SaveMember(store.Member)

	return o
}
