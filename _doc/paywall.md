# Paywall

## Default Paywall Data

    GET /paywall/default

The data is used to render web page for subscripiton

### Response

It should have not other state other than `200 OK`

```json
{
    "banner": {
        "heading": "FT中文网会员订阅服务",
        "subHeading": "欢迎您",
        "coverUrl": "http://www.ftacademy.cn/subscription.jpg",
        "content": [
            "希望全球视野的FT中文网，能够带您站在高海拔的地方俯瞰世界，引发您的思考，从不同的角度看到不一样的事物，见他人之未见！"
        ]
    },
    "products": [
        {
            "heading": "标准会员",
            "benefits": [
                "专享订阅内容每日仅需0.7元(或按月订阅每日0.9元)",
                "精选深度分析",
                "中英双语内容",
                "金融英语速读训练",
                "英语原声电台",
                "无限浏览7日前所有历史文章（近8万篇）"
            ],
            "smallPrint": null,
            "tier": "standard",
            "currency": "CNY",
            "pricing": [
                {
                    "tier": "standard",
                    "cycle": "year",
                    "listPrice": 258,
                    "netPrice": 258,
                    "description": "FT中文网 - 年度标准会员"
                },
                {
                    "tier": "standard",
                    "cycle": "month",
                    "listPrice": 28,
                    "netPrice": 28,
                    "description": "FT中文网 - 月度标准会员"
                }
            ]
        },
        {
            "heading": "高端会员",
            "benefits": [
                "专享订阅内容每日仅需5.5元",
                "享受“标准会员”所有权益",
                "编辑精选，总编/各版块主编每周五为您推荐本周必读资讯，分享他们的思考与观点",
                "FT中文网2018年度论坛门票2张，价值3999元/张 （不含差旅与食宿）"
            ],
            "smallPrint": "注：所有活动门票不可折算现金、不能转让、不含差旅与食宿",
            "tier": "premium",
            "currency": "CNY",
            "pricing": [
                {
                    "tier": "premium",
                    "cycle": "year",
                    "listPrice": 1998,
                    "netPrice": 1998,
                    "description": "FT中文网 - 高端会员"
                }
            ]
        }
    ]
}
```

## Default Pricing Plans

    GET /paywall/pricing/default

```json
{
    "premium_year": {
        "tier": "premium",
        "cycle": "year",
        "listPrice": 1998,
        "netPrice": 1998,
        "description": "FT中文网 - 高端会员"
    },
    "standard_month": {
        "tier": "standard",
        "cycle": "month",
        "listPrice": 28,
        "netPrice": 28,
        "description": "FT中文网 - 月度标准会员"
    },
    "standard_year": {
        "tier": "standard",
        "cycle": "year",
        "listPrice": 258,
        "netPrice": 258,
        "description": "FT中文网 - 年度标准会员"
    }
}
```

## Pricing Plan for Promotion

The data are similar to `/paywall/pricing/default` except the `netPrice` field, which is the actual price user will pay. If a promotion plan exist, it will be set to promotion's net price; if the API is run in `sandbox` mode, `netPrice` is always `0.01`.

```json
{
    "premium_year": {
        "tier": "premium",
        "cycle": "year",
        "listPrice": 1998,
        "netPrice": 0.01,
        "description": "FT中文网 - 高端会员"
    },
    "standard_month": {
        "tier": "standard",
        "cycle": "month",
        "listPrice": 28,
        "netPrice": 0.01,
        "description": "FT中文网 - 月度标准会员"
    },
    "standard_year": {
        "tier": "standard",
        "cycle": "year",
        "listPrice": 258,
        "netPrice": 0.01,
        "description": "FT中文网 - 年度标准会员"
    }
}
```