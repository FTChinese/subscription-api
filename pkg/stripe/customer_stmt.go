package stripe

const StmtSetCustomerID = `
UPDATE cmstmp01.userinfo
SET stripe_customer_id = :stripe_id
WHERE user_id = :ftc_id
LIMIT 1`
