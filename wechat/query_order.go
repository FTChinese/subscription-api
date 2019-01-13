package wechat

// QueryOrder is the query result of wxpay.
type QueryOrder struct {
	OpenID           string `json:"openId"`
	TradeType        string `json:"tradeType"`
	PaymentState     string `json:"paymentState"` // SUCCESS, REFUND, NOTPAY, CLOSED, REVOKED, USERPAYING, PAYERROR
	TotalFee         string `json:"totalFee"`
	TransactionID    string `json:"transactionId"`
	FTCOrderID       string `json:"ftcOrderId"`
	PaidAt           string `json:"paidAt"`
	PaymentStateDesc string `json:"paymentStateDesc"`
}
