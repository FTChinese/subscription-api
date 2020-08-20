package letter

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

type CtxSubs struct {
	UserName string
	Order    subs.Order
}

type CtxUpgrade struct {
	UserName string
	Order    subs.Order
	Prorated []subs.ProratedOrder
}

type CtxIAPLinked struct {
	UserName   string
	Email      string
	Tier       enum.Tier
	ExpireDate chrono.Date
}
