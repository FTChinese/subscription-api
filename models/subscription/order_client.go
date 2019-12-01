package subscription

import "gitlab.com/ftchinese/subscription-api/models/util"

type OrderClient struct {
	OrderID string `db:"order_id"`
	util.ClientApp
}
