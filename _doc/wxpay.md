# Wechat Pay

## Wxpay Unified Order

    POST /wxpay/unified-order/{tier}/{cycle}

Request server to create wxpay's unified order.

### Input

`tier` must be one of `standard` or `premium`.

`cycle` must be one of `year` or `month`.

Request header must contain:
```
X-User-Id: user-uuid
X-Client-Type: <web|ios|android>
X-Client-Version: <major.minor.patch>
```

If `X-Client-Type` is `web`, the request header must also contain
```
X-User-Agent: <forwarded user agent for web> 
X-User-Ip: <forwareded user ip>
```

### Response

* `401 Unauthorized` if request header does not contain `X-User-Id`.

* `400 Bad Request`

If `tier` and `cycle` is not one of the values as specified above;

If subscription plan if not found;

If wechat server send back error;

* `403 Forbidden`

If this user is already a member and current date is not within the allowed renewal period.

```json
{
    "message": "Already a subscribed user and not within allowed renewal period.",
}
```
* `422 Unprocessable Entity`

if `return_code` in wechat's reponse is `FAIL`:
```json
{
    "message": "appid不存在 | 商户号mch_id与appid不匹配 | invalid spbill_create_ip | spbill_create_ip参数长度有误",
    "error": {
        "field": "return_code",
        "code": "fail"
    }
}
```

if `result_code` from wechat is `FAIL`:
```json
{
    "message": "系统异常，请用相同参数重新调用",
    "error": {
        "field": "result_code",
        "code": "SYSTEMERROR"
    }
}
```

* `500 Internal Server Error` for any server or database error.

* `200 OK`

All fields except `ftcOrderId` is required by https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_12&index=2.

`ftcOrderId` is added after all other fields are signed. Client can use it to query order status after payment.

```json
{
    "appid": "wx app id",
    "partnerid": "wx mch id",
    "prepayid": "wx created prepay id",
    "package": "Sign=WXPay",
    "noncestr": "string",
    "timestamp": "unix timestamp",
    "sign": "string",
    "ftcOrderId": "custom field"
}
```

## Wx Query Order

    GET /wxpay/query/{orderId}

### Response

* `401 Unauthorized` if request header does not contain `X-User-Id`.

* `400 Bad Request` if orderId is empty;

* `404 Not Found` if the orderId is not found, or if this order's `appid` and `mchid` is not us.

* `422 Unprocessable Entity`

if wechat's `return_code` is `FAIL`:
```json
{
    "message": "签名失败",
    "error": {
        "field": "return_code",
        "code": "fail"
    }
}
```

if wechat's `result_code` is `FAIL`:
```json
{
    "message": "系统错误",
    "error": {
        "field": "result_code",
        "code": "fail"
    }
}
```

* `500 Internal Server Error` if errors occurred while contacting wechat server.

* `200 OK`
```json
{
    "openId": "string",
    "tradeType": "APP",
    "paymentState": "SUCCESS | REFUND | NOTPAY | CLOSED | REVOKED | USERPAYING | PAYERROR",
    "totalFee": "in cent",
    "transactionId": "string",
    "ftcOrderId": "string",
    "paidAt": "iso8601 time string",
    "paymentStateDesc": "支付失败，请重新下单支付"
}
```

## Notification

Example wx notification data as described on https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_7&index=3

```json
{
    "cash_fee":1,
    "is_subscribe": "N",
    "mch_id": "1504993271",
    "time_end": "20181027175113",
    "total_fee":1,
    "bank_type": "CFT",
    "nonce_str":"1540633845456125000", 
    "result_code":"SUCCESS",
    "return_code":"SUCCESS",
    "transaction_id":"4200000190201810278529489604",
    "fee_type": "CNY",
    "out_trade_no":"FT0055501540633845",
    "sign":"8C4B3D90F2B989EAAC4B541329EF5F8B",
    "appid":"wxacddf1c20516eb69",
    "openid":"ob7fA0h69OO0sTLyQQpYc55iF_P0",
    "trade_type":"APP"
}
```

Order query response:
```json
{
    "return_msg":"OK",
    "total_fee":1,
    "fee_type":"CNY", 
    "transaction_id":"4200000192201810276298895392",
    "out_trade_no":"FT0059051540637408",
    "mch_id":"1504993271",
    "is_subscribe":"N",
    "trade_state_desc":"支付成功",
    "return_code":"SUCCESS",
    "openid":"ob7fA0h69OO0sTLyQQpYc55iF_P0",
    "trade_type":"APP",
    "bank_type":"CFT",
    "trade_state":"SUCCESS",
    "time_end":"20181027185023",
    "cash_fee":1,
    "appid":"wxacddf1c20516eb69",
    "nonce_str":"iP1joB3m1zqDAzUL",
    "sign":"2C5F0583127CE3D975DEC5EF2E4A8C45",
    "result_code":"SUCCESS",
    "attach":""
}
```