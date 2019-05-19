package test

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/util"
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

func (m Model) ClearFTCMember(id string) error {
	query := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id = ?`

	_, err := m.db.Exec(query, id)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) ClearWxMember(unionID string) error {
	query := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id_alias = ?`

	_, err := m.db.Exec(query, unionID)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) ClearUser(id string) error {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE user_id = ?`

	_, err := m.db.Exec(query, id)

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

func (m Model) CreateUser(u paywall.FtcUser, app util.ClientApp) error {
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

	_, err := m.db.Exec(query,
		u.UserID,
		u.UnionID,
		u.Email,
		"12345678",
		u.UserName,
		app.ClientType,
		app.Version,
		app.UserIP,
		app.UserAgent)

	if err != nil {
		return err
	}

	return nil
}

func (m Model) CreateMember(mm paywall.Membership) error {
	_, err := m.db.Exec(
		m.query.UpsertMember(),
		mm.CompoundID,
		mm.UnionID,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate,
		mm.FTCUserID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		mm.ExpireDate)

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
