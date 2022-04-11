//go:build !production
// +build !production

package price

import "github.com/FTChinese/go-rest/enum"

var MockEditionStdYear = Edition{
	Tier:  enum.TierStandard,
	Cycle: enum.CycleYear,
}

var MockEditionStdMonth = Edition{
	Tier:  enum.TierStandard,
	Cycle: enum.CycleMonth,
}

var MockEditionPrm = Edition{
	Tier:  enum.TierPremium,
	Cycle: enum.CycleYear,
}
