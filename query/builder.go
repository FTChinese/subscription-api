package query

const (
	selectMember = `
	SELECT vip_id AS userId,
		vip_id_alias AS unionId,
		CASE vip_type
			WHEN 10 THEN 'standard'
			WHEN 100 THEN 'premium'
			ELSE member_tier
		END AS memberTier,
		billing_cycle AS billingCycle,
		IF(expire_time, DATE(FROM_UNIXTIME(expire_time)), expire_date) AS expireDate
	FROM %s.ftc_vip
	WHERE vip_id = ? 
		OR vip_id_alias = ?
	LIMIT 1`
)

type Builder struct {
	sandbox bool
}

func (b Builder) MemberDB() string {
	if b.sandbox {
		return "sandbox"
	}

	return "premium"
}

func NewBuilder(sandbox bool) Builder {
	return Builder{sandbox: sandbox}
}
