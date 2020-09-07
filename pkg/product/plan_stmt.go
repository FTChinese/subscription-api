package product

const colExpandedPlan = `
SELECT p.id AS plan_id,
	p.product_id,
	p.product_id,
	p.price,
	p.tier,
	p.cycle,
	p.description,
	d.id AS discount_id,
	d.price_off,
	d.percent,
	d.start_utc,
	d.end_utc,
	d.description AS discount_desc
`

// StmtExpandedPlanByID retrieves a single plan with discount by plan id.
const StmtExpandedPlanByID = colExpandedPlan + `
FROM subs_product.plan AS p
	LEFT JOIN subs_product.discount AS d
	ON p.discount_id = d.id
WHERE p.id = ?
LIMIT 1`

const StmtExpandedPlanByEdition = colExpandedPlan + `
FROM subs_product.product_active_plans AS a
	LEFT JOIN subs_product.plan AS p
		ON a.plan_id = p.id
	LEFT JOIN subs_product.discount AS d
		ON p.discount_id = d.id
WHERE p.tier = ? AND p.cycle = ?
LIMIT 1`

// StmtPaywallPlans selects all active plans of products which are listed on paywall.
// The plans has discount attached.
const StmtPaywallPlans = colExpandedPlan + `
FROM subs_product.product_active_plans AS a
	LEFT JOIN subs_product.plan AS p
		ON a.plan_id = p.id
	LEFT JOIN subs_product.discount AS d
		ON p.discount_id = d.id
	LEFT JOIN subs_product.paywall_product AS pp
		ON a.product_id = pp.product_id
WHERE p.id IS NOT NULL
	AND pp.product_id IS NOT NULL
ORDER BY cycle DESC`
