package model

import (
	"database/sql"
	"fmt"

	"github.com/FTChinese/go-rest/enum"

	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// Env wraps database connection
type Env struct {
	DB    *sql.DB
	Cache *cache.Cache
}

var logger = log.WithField("package", "model")

const (
	keySchedule = "discountSchedule"
	keyPromo    = "promotionSchedule"

	stmtSubs = `
	SELECT user_id AS userId,
		trade_no AS orderId,
		trade_price AS price,
		trade_amount AS charged,
		login_method AS loginMethod,
		tier_to_buy AS tierToBuy,
		billing_cycle AS billingCycle,
		payment_method AS paymentMethod,
		created_utc AS createdAt,
		confirmed_utc AS confirmedAt
	FROM premium.ftc_trade
	WHERE trade_no = ?
	LIMIT 1`

	stmtUpdateSubs = `
	UPDATE premium.ftc_trade
	SET is_renewal = ?,
		confirmed_utc = ?,
		start_date = ?,
		end_date = ?
	WHERE trade_no = ?
	LIMIT 1`

	stmtCreateMember = `
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
)

var (
	stmtSubsLock = fmt.Sprintf(`%s
	FOR UPDATE`, stmtSubs)
)

func normalizeMemberTier(vipType int64) enum.Tier {
	switch vipType {

	case 10:
		return enum.TierStandard

	case 100:
		return enum.TierPremium

	default:
		return enum.InvalidTier
	}
}
