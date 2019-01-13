package wechat

// PrepayOrder is the response send to client
type PrepayOrder struct {
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
