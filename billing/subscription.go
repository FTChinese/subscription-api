package billing

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

type Subscription struct {
	ID            string         `json:"id"`
	FtcID         string         `json:"ftcId"`
	Plan          Plan           `json:"plan"`
	Product       Product        `json:"product"`
	Coupon        Coupon         `json:"coupon"`
	Upgrade       Upgrade        `json:"upgrade,omitempty"`
	Usage         string         `json:"usage"`
	PaymentMethod enum.PayMethod `json:"paymentMethod"`
	WxAppID       null.String    `json:"-"` // Wechat specific
	CreatedAt     chrono.Time    `json:"createdAt"`
	ConfirmedAt   chrono.Time    `json:"confirmedAt"` // When the payment is confirmed.
	StartDate     chrono.Date    `json:"startDate"`   // Membership start date for this order. If might be ConfirmedAt or user's existing membership's expire date.
	EndDate       chrono.Date    `json:"endDate"`     // Membership end date for this order. Depends on start date.
}
