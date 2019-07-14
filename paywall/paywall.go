package paywall

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// Banner is the banner used on the barrier page
type Banner struct {
	Heading    string   `json:"heading"`
	SubHeading string   `json:"subHeading"`
	CoverURL   string   `json:"coverUrl"`
	Content    []string `json:"content"`
}

// ProductCard contains data to show the description of a subscription product.
type ProductCard struct {
	Heading    string      `json:"heading"`
	Benefits   []string    `json:"benefits"`
	SmallPrint null.String `json:"smallPrint"`
	Tier       enum.Tier   `json:"tier"`
	Currency   string      `json:"currency"`
	Pricing    []Plan      `json:"pricing"`
}

type PayWall struct {
	Banner   Banner        `json:"banner"`
	Products []ProductCard `json:"products"`
}

// BuildPayWall constructs the data used to show pay wall.
func BuildPayWall(banner Banner, pricing FtcPlans) (PayWall, error) {
	planStdYear, err := pricing.FindPlan(
		enum.TierStandard.String(),
		enum.CycleYear.String())
	if err != nil {
		return PayWall{}, err
	}

	planStdMonth, err := pricing.FindPlan(
		enum.TierStandard.String(),
		enum.CycleMonth.String())
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
		Products: []ProductCard{
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
