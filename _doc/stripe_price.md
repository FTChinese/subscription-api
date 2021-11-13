## 获取Stripe价格列表

```
GET /stripe/prices?refresh=true|false
```

改请求带有一个可选参数 `refresh`，值为字符串 `true` 或 `false`。用户强制从Stripe API获取数据，否则返回可能是缓存数据。

### Workflow

1. 解析URL的表单参数。如果无法解析，返回 `400 Bad Request`。仅接受字符串`"true"` 为 `true`，其他所有值被当作 `false`。
2. 如果 `refresh` 不是 `true`，且缓存中存在数据， 则返回缓存数据
3. 否则，从 Stripe [List all prices](https://stripe.com/docs/api/prices/list) 获取数据
4. 缓存数据
5. 返回数据

### Example response:

```json
[
    {
        "id": "plan_FOdgPTznDwHU4i",
        "tier": "standard",
        "cycle": "month",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Standard Monthly Plan",
        "productId": "prod_FOde1wE4ZTRMcD",
        "unitAmount": 390,
        "created": 1562567567
    },
    {
        "id": "plan_FOdfeaqzczp6Ag",
        "tier": "standard",
        "cycle": "year",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Standard Yearly Plan",
        "productId": "prod_FOde1wE4ZTRMcD",
        "unitAmount": 3000,
        "created": 1562567504
    },
    {
        "id": "plan_FOde0uAr0V4WmT",
        "tier": "premium",
        "cycle": "year",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Premium Yearly Plan",
        "productId": "prod_FOdd1iNT29BIGq",
        "unitAmount": 23800,
        "created": 1562567431
    }
]
```
