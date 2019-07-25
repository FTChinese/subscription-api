package test

import (
	"database/sql"
	"github.com/guregu/null"
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

func (m Model) ClearUser(u paywall.AccountID) error {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE user_id = ?
		OR wx_union_id = ?
	LIMIT 1`

	_, err := m.db.Exec(query, u.CompoundID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) ClearMember(u paywall.AccountID) error {
	q := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id = ?
		OR vip_id_alias = ?`

	_, err := m.db.Exec(q, u.CompoundID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) ClearOrder(u paywall.AccountID) error {
	q := `
	DELETE FROM premium.ftc_trade
	WHERE user_id IN (?, ?)`

	_, err := m.db.Exec(q, u.FtcID, u.UnionID)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) ClearUserByEmail(email string) error {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE email = ?`

	_, err := m.db.Exec(query, email)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) CreateFtcUser(p Profile) error {
	query := `
	INSERT INTO cmstmp01.userinfo
	SET user_id = ?,
		wx_union_id = ?,
		email = ?,
		password = MD5(?),
		user_name = ?,
		client_type = ?,
		client_version = ?,
		user_ip = INET6_ATON(?),
		user_agent = ?,
		created_utc = UTC_TIMESTAMP()`

	app := RandomClientApp()

	_, err := m.db.Exec(query,
		p.FtcID,
		null.String{},
		p.Email,
		p.Password,
		p.UserName,
		app.ClientType,
		app.Version,
		app.UserIP,
		app.UserAgent)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) CreateGiftCard() paywall.GiftCard {
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
