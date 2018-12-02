# Paywall

## Promotion Schedule

    GET /paywall/promo

Get a promotion schedule from cache.

You must hit the `/__refresh` endpoint before having this one to take effect.

### Response
`200 OK`
```json
{
    "startAt": "2018-12-31T16:00:00Z",
    "endAt": "2019-01-01T16:00:00Z",
    "plans": {
        "standard_month": {
            "tier": "standard",
            "cycle": "month",
            "price": 28,
            "id": 5,
            "description": "FT中文网 - 标准会员"
        },
        "standard_year": {
            "tier": "standard",
            "cycle": "year",
            "price": 168,
            "id": 10,
            "description": "FT中文网 - 标准会员"
        },
        "premium_year": {
            "tier": "premium",
            "cycle": "year",
            "price": 1698,
            "id": 100,
            "description": "FT中文网 - 高端会员"
        }
    },
    "banner": {
        "coverUrl": "https://cn.bing.com/az/hprichbg/rb/FrankfurtXmas_ZH-CN9289866662_1920x1080.jpg",
        "heading": "FT中文网会员订阅服务",
        "subHeading": "欢迎您",
        "content": [
            "希望全球视野的FT中文网，能够带您站在高海拔的地方俯瞰世界，引发您的思考，从不同的角度看到不一样的事物，见他人之未见！",
            "If you need to display multiple paragraphs, press Enter at the end of each paragraph.",
            "Do not add empty lines between paragraphs; otherwise the empty paragraph will also be displayed."
        ]
    },
    "createdAt": "2018-11-29T08:53:00Z"
}
```

`404 Not Found` if no promotion is scheduled, e.g., you never hit the `/__refresh` endpoint.

## Default Products Description

    GET /paywall/products

Not implemented yet.

## Default Banner on Barrier Page

    GET /paywall/banner

Not implemented yet.

## Default Pricing Plans

    GET /paywall/plans

This is the canonical pricing plans you would use on a daily basis. If the promotion endpoint is in effect (depending on the start and end date), use the promotion schedule's `plans` field; otherwise use this one.

```json
{
    "premium_year": {
        "tier": "premium",
        "cycle": "year",
        "price": 1998,
        "id": 100,
        "description": "FT中文网 - 高端会员"
    },
    "standard_month": {
        "tier": "standard",
        "cycle": "month",
        "price": 28,
        "id": 5,
        "description": "FT中文网 - 月度标准会员"
    },
    "standard_year": {
        "tier": "standard",
        "cycle": "year",
        "price": 198,
        "id": 10,
        "description": "FT中文网 - 年度标准会员"
    }
}
```