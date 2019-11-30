package subscription

import "github.com/FTChinese/go-rest/enum"

type PlanBase struct {
	Product enum.Tier  `json:"tier"`
	Cycle   enum.Cycle `json:"cycle"`
}

func (p PlanBase) ProductKey() string {
	return p.Product.String() + "_" + p.Cycle.String()
}

type ProductID struct {
	StripeTest string
	StripeLive string
	Apple      string
}

type Plan struct {
	PlanBase
	ListPrice  float64   `json:"listPrice" db:"price"`
	NetPrice   float64   `json:"netPrice" db:"amount"`
	CycleCount int64     `json:"cycleCount" db:"cycle_count"`
	ExtraDays  int64     `json:"extraDays" db:"extra_days"`
	Currency   string    `json:"currency" db:"currency"`
	Title      string    `json:"description" db:"title"`
	ProductID  ProductID `json:"-"`
}
