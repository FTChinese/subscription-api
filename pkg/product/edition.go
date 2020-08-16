package product

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

type Edition struct {
	Tier  enum.Tier  `json:"tier" db:"tier"`
	Cycle enum.Cycle `json:"cycle" db:"cycle"`
}

func (e Edition) PaymentTitle(k enum.OrderKind) string {
	return fmt.Sprintf("%sFT中文网%s/%s", k.StringSC(), e.Tier.StringCN(), e.Cycle.StringCN())
}

func (e Edition) NamedKey() string {
	return e.Tier.String() + "_" + e.Cycle.String()
}

// String produces a human readable string of this edition.
// * 标准会员/年
// * 标准会员/月
// * 高端会员/年
func (e Edition) String() string {
	return e.Tier.StringCN() + "/" + e.Cycle.StringCN()
}
