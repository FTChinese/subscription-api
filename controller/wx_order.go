package controller

// WxConfig contains wxpay configuration.
type WxConfig struct {
	AppID  string
	MchID  string
	APIKey string
	IsProd bool
}

// WxOrder is the response send to client which is requesting to pay by tenpay
type WxOrder struct {
	FtcOrderID string  `json:"ftcOrderId"`
	Price      float64 `json:"price"`
	AppID      string  `json:"appid"`
	PartnerID  string  `json:"partnerid"`
	PrepayID   string  `json:"prepayid"`
	Package    string  `json:"package"`
	Nonce      string  `json:"noncestr"`
	Timestamp  string  `json:"timestamp"`
	Signature  string  `json:"sign"`
}

// WxQueryOrder is the query result of wxpay.
type WxQueryOrder struct {
	OpenID           string `json:"openId"`
	TradeType        string `json:"tradeType"`
	PaymentState     string `json:"paymentState"` // SUCCESS, REFUND, NOTPAY, CLOSED, REVOKED, USERPAYING, PAYERROR
	TotalFee         string `json:"totalFee"`
	TransactionID    string `json:"transactionId"`
	FTCOrderID       string `json:"ftcOrderId"`
	PaidAt           string `json:"paidAt"`
	PaymentStateDesc string `json:"paymentStateDesc"`
}
