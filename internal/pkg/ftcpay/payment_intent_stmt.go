package ftcpay

const StmtSavePaymentIntent = `
INSERT INTO premium.ftc_payment_intent
SET order_id = :order_id,
	price = :price,
	offer = :offer,
	membership = :membership,
	alipay_params = :alipay_params,
	wxpay_params = :wxpay_params,
	created_utc = :created_utc
`

const StmtRetrievePaymentIntent = `
SELECT order_id,
	price,
	offer,
	membership
FROM premium.ftc_payment_intent
WHERE order_id = ?
LIMIT 1
`
