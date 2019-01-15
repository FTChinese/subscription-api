package ali

// AppPayResp is the response send to client which is requesting to pay by alipay.
type AppPayResp struct {
	FtcOrderID string  `json:"ftcOrderId"`
	Price      float64 `json:"price"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	Param      string  `json:"param"`
}
