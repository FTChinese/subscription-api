# Subscription API

API for subscription service

## Endpoints

### Place order

* POST `/place-order/alipay/{standard|premium}/{year|month}` Create an order for alipay.

Request header:
```
X-User-Id: user-uuid
X-Client-Type: <web|ios|android>
X-Client-Version: <major.minor.path>
```

`X-User-Agent: <forwarded user agent for web>` required for web app.

Response:
```json
{
    "order": "string"
}
```

* POST `/place-order/tenpay/{standard|premium}/{year|month}`

Request header:
```
X-User-Id: user-uuid
X-Client-Type: <web|ios|android>
X-Client-Version: <major.minor.path>
```

`X-User-Agent: <forwarded user agent for web>` required for web app.

Response:
```json
{
    "appid": "",
    "partnerid": "",
    "prepayid": "",
    "package": "Sign=WXPay",
    "noncestr": "",
    "timestamp": "",
    "sign": "",
}
```

* POST `/verify/alipay/app-pay`

* POST `/notify/alipay`

* POST `/notify/tenpay`

* `GET /query-order/tenpay/{orderId}`