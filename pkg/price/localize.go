package price

import (
	"strconv"
	"strings"

	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
)

var tiersCN = map[enum.Tier]string{
	enum.TierStandard: "标准会员",
	enum.TierPremium:  "高端会员",
}

var cyclesCN = map[enum.Cycle]string{
	enum.CycleYear:  "年",
	enum.CycleMonth: "月",
}

var orderKindsCN = map[enum.OrderKind]string{
	enum.OrderKindCreate:  "订阅",
	enum.OrderKindRenew:   "续订",
	enum.OrderKindUpgrade: "升级订阅",
	enum.OrderKindAddOn:   "购买",
}

func LocalizeTierCN(t enum.Tier) string {
	return tiersCN[t]
}

func LocalizeCycleCN(c enum.Cycle) string {
	return cyclesCN[c]
}

// GetPerCycle returns a human-readable string of every year or month.
// - /年
// - /月
func GetPerCycle(c enum.Cycle) string {
	return "/" + cyclesCN[c]
}

// LocalizeEditionCN produces a human-readable string of an edition.
// - 标准会员/年
// - 标准会员/月
// - 高端会员/年
func LocalizeEditionCN(e Edition) string {
	return tiersCN[e.Tier] + "/" + cyclesCN[e.Cycle]
}

func LocalizePeriodCN(ymd dt.YearMonthDay) string {
	var b strings.Builder

	if ymd.Years > 0 {
		b.WriteString(strconv.FormatInt(ymd.Years, 10))
		b.WriteString("年")
	}

	if ymd.Months > 0 {
		b.WriteString(strconv.FormatInt(ymd.Months, 10))
		b.WriteString("月")
	}

	if ymd.Days > 0 {
		b.WriteString(strconv.FormatInt(ymd.Days, 10))
		b.WriteString("天")
	}

	return b.String()
}

func LocalizeTierPeriodCN(t enum.Tier, p dt.YearMonthDay) string {
	var b strings.Builder

	b.WriteString(tiersCN[t])
	b.WriteByte('/')

	if p.IsSingular() {
		b.WriteString(cyclesCN[p.EqCycle()])
	} else {
		b.WriteString(LocalizePeriodCN(p))
	}

	return b.String()
}

func BuildPaymentTitle(k enum.OrderKind, t enum.Tier, p dt.YearMonthDay) string {
	return "FT中文网" + LocalizeTierPeriodCN(t, p) + " - " + orderKindsCN[k]
}
