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

func NewModel(db *sql.DB) Model {
	return Model{
		db:    db,
		query: query.NewBuilder(false),
	}
}

func (model Model) ClearUser(u paywall.AccountID) error {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE user_id = ?
		OR wx_union_id = ?
	LIMIT 1`

	_, err := model.db.Exec(query, u.CompoundID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (model Model) ClearMember(u paywall.AccountID) error {
	q := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id = ?
		OR vip_id_alias = ?`

	_, err := model.db.Exec(q, u.CompoundID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (model Model) ClearOrder(u paywall.AccountID) error {
	q := `
	DELETE FROM premium.ftc_trade
	WHERE user_id IN (?, ?)`

	_, err := model.db.Exec(q, u.FtcID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (model Model) ClearUserByEmail(email string) error {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE email = ?`

	_, err := model.db.Exec(query, email)

	if err != nil {
		return err
	}

	return nil
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

func (model Model) CreateSub(s paywall.Subscription) {

	var stmt = model.query.InsertSubs() + `
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?`

	c := RandomClientApp()

	_, err := model.db.Exec(stmt,
		s.ID,
		s.CompoundID,
		s.FtcID,
		s.UnionID,
		s.ListPrice,
		s.NetPrice,
		s.Tier,
		s.Cycle,
		s.CycleCount,
		s.ExtraDays,
		s.Usage,
		s.PaymentMethod,
		s.WxAppID,
		s.CreatedAt,
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

func (model Model) CreateMember(m paywall.Membership) {
	_, err := model.db.Exec(model.query.InsertMember(),
		m.ID,
		m.CompoundID,
		m.UnionID,
		m.TierCode(),
		m.ExpireDate.Unix(),
		m.FtcID,
		m.UnionID,
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
