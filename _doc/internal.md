# Server Internal Status

## Cache Latest Promotion Schedule

    GET /__refresh

Tell this API to retrieve a promotion schedule and put it into cache. It will only retrieve one row from database whose creation time is the latest. You have no way to tell it which only to pick.

### Response

`200 OK`:
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

Usually it would always returns a promotion schedule, even though the promotion is already expired, unless there is no single row in database.

## See Current Effective Pricing Plans

    GET /__current_plans

This is used to see what prices will be used to charge users when user clicked the pay button on client apps.

### Response

Always `200 OK`
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

If a promotion schedule is in effect, the prices might be different. You can use those data to see if we are charging users with wrong prices.

## Build Version

    GET /__version

The running program's build verison, commit tag and when it is built.

```json
{
    "build": "2018-12-02T11:13:33+0800",
    "version": "v0.1.0-24-g02ea926"
}
```