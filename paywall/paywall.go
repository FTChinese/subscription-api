package paywall

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Banner is the banner used on the barrier page
type Banner struct {
	Heading    string   `json:"heading"`
	CoverURL   string   `json:"coverUrl"`
	SubHeading string   `json:"subHeading"`
	Content    []string `json:"content"`
}

// Plan is a pricing plan.
// The list price is the price that buyers pay for your product or service without any discounts.
// The net price of a product or service is the actual price that customers pay for the product or service.
type Plan struct {
	Tier        enum.Tier  `json:"tier"`
	Cycle       enum.Cycle `json:"cycle"`
	ListPrice   float64    `json:"listPrice"`
	NetPrice    float64    `json:"netPrice"`
	Description string     `json:"description"`
}

// Product contains data to show the description of a subscription product.
type Product struct {
	Heading    string      `json:"heading"`
	Benefits   []string    `json:"benefits"`
	SmallPrint null.String `json:"smallPrint"`
	Tier       enum.Tier   `json:"tier"`
	Currency   string      `json:"currency"`
	Pricing    []Plan      `json:"pricing"`
}

type PayWall struct {
	Banner   Banner    `json:"banner"`
	Products []Product `json:"products"`
}

// BuildPayWall constructs the data used to show pay wall.
func BuildPayWall(banner Banner, pricing Pricing) (PayWall, error) {
	planStdYear, err := pricing.FindPlan(
		enum.TierStandard.String(),
		enum.CycleYear.String())
	if err != nil {
		return PayWall{}, err
	}

	planStdMonth, err := pricing.FindPlan(
		enum.TierStandard.String(),
		enum.CycleYear.String())
	if err != nil {
		return PayWall{}, err
	}

	planPrmYear, err := pricing.FindPlan(
		enum.TierPremium.String(),
		enum.CycleYear.String())
	if err != nil {
		return PayWall{}, err
	}

	return PayWall{
		Banner: banner,
		Products: []Product{
			{
				Heading: "标准会员",
				Benefits: []string{
					"专享订阅内容每日仅需0.7元(或按月订阅每日0.9元)",
					"精选深度分析",
					"中英双语内容",
					"金融英语速读训练",
					"英语原声电台",
					"无限浏览7日前所有历史文章（近8万篇）",
				},
				SmallPrint: null.String{},
				Tier:       enum.TierStandard,
				Currency:   "CNY",
				Pricing: []Plan{
					planStdYear,
					planStdMonth,
				},
			},
			{
				Heading: "高端会员",
				Benefits: []string{
					"专享订阅内容每日仅需5.5元",
					"享受“标准会员”所有权益",
					"编辑精选，总编/各版块主编每周五为您推荐本周必读资讯，分享他们的思考与观点",
					"FT中文网2018年度论坛门票2张，价值3999元/张 （不含差旅与食宿）",
				},
				SmallPrint: null.StringFrom("注：所有活动门票不可折算现金、不能转让、不含差旅与食宿"),
				Tier:       enum.TierPremium,
				Currency:   "CNY",
				Pricing: []Plan{
					planPrmYear,
				},
			},
		},
	}, nil
}
