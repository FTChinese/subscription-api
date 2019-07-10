package billing

import "github.com/FTChinese/go-rest/chrono"

type Coupon struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	AmountOff float64     `json:"amountOff"`
	CreatedAt chrono.Time `json:"createdAt"`
	StartUTC  chrono.Time `json:"startAt"`
	EndUTC    chrono.Time `json:"endAt"`
}
