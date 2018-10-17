# Subscription API

API for subscription service

## Endpoints

### Place order

* POST `/order/alipay/<standard|premium>/<year|month|week>` Create an order for alipay.

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

* POST `/order/tenpay/<standard|premium>/<year|month|week>`

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

* POST `/notify/alipay`

* POST `/notify/tenpay`

* POST `/verify/alipay/app-pay`