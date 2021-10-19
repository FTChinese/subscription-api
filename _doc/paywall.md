# Paywall

## Default Paywall Data

    GET /paywall/default

The data is used to render web page for subscripiton

### Response

It should have not other state other than `200 OK`

```json
{
  "banner": {
    "id": 1,
    "heading": "FT中文网会员订阅服务",
    "subHeading": "欢迎您！",
    "coverUrl": "http://www.ftacademy.cn/subscription.jpg",
    "content": "希望全球视野的FT中文网，能够带您站在高海拔的地方俯瞰世界，引发您的思考，从不同的角度看到不一样的事物，见他人之未见！"
  },
  "promo": {
    "id": null,
    "heading": null,
    "subHeading": null,
    "coverUrl": null,
    "content": null,
    "terms": null,
    "startUtc": null,
    "endUtc": null
  },
  "products": [
    {
      "id": "prod_9xrJdHFq0wmq",
      "tier": "standard",
      "heading": "标准会员",
      "description": "专享订阅内容每日仅需0.83元(或按月订阅每日1.17元)\n精选深度分析\n中英双语内容\n金融英语速读训练\n英语原声电台\n阅读1日前历史文章（近9万篇）",
      "smallPrint": null,
      "isActive": false,
      "createdUtc": "2021-01-25T01:51:09Z",
      "updatedUtc": "2021-10-08T02:32:31Z",
      "createdBy": "weiguo.ni",
      "prices": [
        {
          "id": "plan_RKy1IuKSXyua",
          "tier": "standard",
          "cycle": "year",
          "active": true,
          "archived": false,
          "currency": "cny",
          "description": "Standard Yearly Edition",
          "liveMode": true,
          "nickname": null,
          "productId": "prod_9xrJdHFq0wmq",
          "source": "ftc",
          "unitAmount": 298,
          "createdUtc": "2021-02-18T02:44:55Z",
          "createdBy": "weiguo.ni",
          "offers": [
            {
              "id": "dsc_fAFaUX9VreYX",
              "createdBy": "weiguo.ni",
              "description": "现在续订享75折优惠",
              "kind": "retention",
              "percent": null,
              "startUtc": null,
              "endUtc": null,
              "priceOff": 80,
              "priceId": "plan_RKy1IuKSXyua",
              "recurring": true,
              "liveMode": true,
              "status": "active",
              "createdUtc": "2021-09-17T07:32:52Z"
            },
            {
              "id": "dsc_2XOFwX42bAvI",
              "createdBy": "weiguo.ni",
              "description": "重新购买会员享85折优惠",
              "kind": "win_back",
              "percent": null,
              "startUtc": null,
              "endUtc": null,
              "priceOff": 40,
              "priceId": "plan_RKy1IuKSXyua",
              "recurring": true,
              "liveMode": true,
              "status": "active",
              "createdUtc": "2021-09-17T07:33:49Z"
            }
          ]
        },
        {
          "id": "plan_ohky3lyEMPSf",
          "tier": "standard",
          "cycle": "month",
          "active": true,
          "archived": false,
          "currency": "cny",
          "description": "Standard Monthly Edition",
          "liveMode": true,
          "nickname": null,
          "productId": "prod_9xrJdHFq0wmq",
          "source": "ftc",
          "unitAmount": 35,
          "createdUtc": "2021-02-18T02:45:18Z",
          "createdBy": "weiguo.ni",
          "offers": []
        }
      ]
    },
    {
      "id": "prod_zSgOTS6DWLmu",
      "tier": "premium",
      "heading": "高端会员",
      "description": "专享订阅内容每日仅需5.5元\n享受“标准会员”所有权益\n编辑精选，总编/各版块主编每周五为您推荐本周必读资讯，分享他们的思考与观点\nFT商学院高端专享\nFT中文网2022年度论坛门票2张",
      "smallPrint": "注：所有活动门票不可折算现金、不能转让、不含差旅与食宿",
      "isActive": false,
      "createdUtc": "2021-01-25T01:48:34Z",
      "updatedUtc": "2021-10-15T05:27:25Z",
      "createdBy": "weiguo.ni",
      "prices": [
        {
          "id": "plan_rLIy6LJYW8LV",
          "tier": "premium",
          "cycle": "year",
          "active": true,
          "archived": false,
          "currency": "cny",
          "description": null,
          "liveMode": true,
          "nickname": null,
          "productId": "prod_zSgOTS6DWLmu",
          "source": "ftc",
          "unitAmount": 1998,
          "createdUtc": "2021-01-25T01:48:34Z",
          "createdBy": "weiguo.ni",
          "offers": [
            {
              "id": "dsc_4jvMFdhEYXWk",
              "createdBy": "weiguo.ni",
              "description": "现在续订享75折优惠",
              "kind": "retention",
              "percent": null,
              "startUtc": null,
              "endUtc": null,
              "priceOff": 500,
              "priceId": "plan_rLIy6LJYW8LV",
              "recurring": true,
              "liveMode": true,
              "status": "active",
              "createdUtc": "2021-09-17T07:35:09Z"
            },
            {
              "id": "dsc_GzsiHfNQuKyn",
              "createdBy": "weiguo.ni",
              "description": "重新购买会员享85折优惠",
              "kind": "win_back",
              "percent": null,
              "startUtc": null,
              "endUtc": null,
              "priceOff": 300,
              "priceId": "plan_rLIy6LJYW8LV",
              "recurring": true,
              "liveMode": true,
              "status": "active",
              "createdUtc": "2021-09-17T07:35:48Z"
            }
          ]
        }
      ]
    }
  ],
  "liveMode": true
}
```




