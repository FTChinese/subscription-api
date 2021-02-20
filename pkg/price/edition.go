package price

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
)

var (
	StdMonthEdition = Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleMonth,
	}

	StdYearEdition = Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleYear,
	}

	PremiumEdition = Edition{
		Tier:  enum.TierPremium,
		Cycle: enum.CycleYear,
	}
)

type Edition struct {
	Tier  enum.Tier  `json:"tier" db:"tier"`
	Cycle enum.Cycle `json:"cycle" db:"cycle"`
}

func (e Edition) Validate() *render.ValidationError {
	if e.Tier == enum.TierNull {
		return &render.ValidationError{
			Message: "Please specify the edition you want to subscribe to",
			Field:   "tier",
			Code:    render.CodeMissingField,
		}
	}

	if e.Cycle == enum.CycleNull {
		return &render.ValidationError{
			Message: "Please specify the billing cycle of your subscription",
			Field:   "cycle",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

func (e Edition) NamedKey() string {
	return e.Tier.String() + "_" + e.Cycle.String()
}

// StringCN produces a human readable string of this edition.
// * 标准会员/年
// * 标准会员/月
// * 高端会员/年
func (e Edition) StringCN() string {
	return e.Tier.StringCN() + "/" + e.Cycle.StringCN()
}
