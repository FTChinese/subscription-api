package controller

// WxOrder is the query result of wxpay.
type WxOrder struct {
	OpenID           string `json:"openId"`
	TradeType        string `json:"tradeType"`
	PaymentState     string `json:"paymentState"`
	TotalFee         string `json:"totalFee"`
	TransactionID    string `json:"transactionId"`
	FTCOrderID       string `json:"ftcOrderId"`
	PaidAt           string `json:"paidAt"`
	PaymentStateDesc string `json:"paymentStateDesc"`
}
