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
		s.NetPrice,
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

// Create a new order and membership.
func (model Model) CreateNewMember(store *SubStore) {
	store.AddOrder(paywall.SubsKindCreate)

	model.SaveSub(store.GetLastOrder())
	model.SaveMember(store.Member)
}

// Create a new order and renew membership.
func (model Model) RenewMember(store *SubStore) {
	store.AddOrder(paywall.SubsKindRenew)

	model.SaveSub(store.GetLastOrder())
	model.UpdateMember(store.Member)
}

// Create n orders to renew a member.
func (model Model) RenewMemberN(store *SubStore, n int) {
	for i := 0; i < n; i++ {
		model.RenewMember(store)
	}
}
