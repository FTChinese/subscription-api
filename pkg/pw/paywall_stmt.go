package pw

const StmtBanner = `
SELECT b.id AS banner_id,
	b.heading AS heading,
	b.cover_url AS cover_url,
	b.sub_heading AS sub_heading,
	b.content AS content,
	p.id AS promo_id,
	p.heading AS promo_heading,
	p.cover_url AS promo_cover_url,
	p.sub_heading AS promo_sub_heading,
	p.content AS promo_content,
	p.terms_conditions AS terms_conditions,
	p.start_utc AS start_utc,
	p.end_utc AS end_utc
FROM subs_product.paywall_banner AS b
	LEFT JOIN subs_product.paywall_promo AS p
	ON b.promo_id = p.id
WHERE b.id = 1
LIMIT 1`
