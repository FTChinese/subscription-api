package query

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

const (
	stmtSelectMember = `
	SELECT id AS member_id, 
		vip_id AS compound_id,
		NULLIF(vip_id, vip_id_alias) AS ftc_id,
		vip_id_alias AS union_id,
		vip_type,
		expire_time,
		member_tier AS tier,
		billing_cycle AS cycle,
		expire_date,
		payment_method,
		stripe_subscription_id AS stripe_sub_id,
		auto_renewal,
		sub_status`
)
