# Wechat Pay

## Shared Input and Output

### Header

`X-User-Id` is FTC account's uuid if user logged in with FTC account while `X-Union-Id` is wechat's union id if user logged in via wechat OAuth. You should provide at least one of them so that user could be identified. If user' FTC account is bound to wechat account, you should provide both of them.

Since this is used only by desktop browsers, the API has no way to know user's ip and user agent. Thus they must be forwarded by the client app.

Native apps do not needs provide `X-User-Ip` and `X-User-Agent` since the request is issued directly from user's device.

### Error Response

* `/wxpay/desktop`
* `/wxpay/mobile`
* `/wxpay/jsapi`
* `/wxpay/app`

The above endpoints have similar error response as follows:

* `401 Unauthorized` 

if request header does not contain either `X-User-Id` nor `X-Union-Id`.

* `400 Bad Request`

If the combination of `tier` and `cycle` is not one of `standard_month`, `standard_year` or `premium_year`. 

If wechat server send back error;

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
    "message": "系统异常，请用相同参数重新调用 | JSAPI支付必须传openid",
    "error": {
        "field": "result_code",
        "code": "SYSTEMERROR | PARAM_ERROR"
    }
}
```

* `500 Internal Server Error` 

For any server or database error.

## In Desktop Browser

    POST /wxpay/desktop/{standard|premium}/{year|month}

This is used to create a QR code in desktop browsers. It corresponds to `trade_type=NATIVE` as specified by wechat. See the official documentation [扫码支付](https://pay.weixin.qq.com/wiki/doc/api/native.php?chapter=6_1)

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```

### Input

None.

### Response

See the [Error Response](#error-response) section for errors.

* `200 OK`

```json
{
    "ftcOrderId": "FTB55FE0DE847779C9",
    "listPrice": 258,
    "netPrice": 0.01,
    "appId": "***REMOVED***",
    "codeUrl": "weixin://wxpay/bizpayurl?pr=CbhlUTz"
}
```

Client should use the `codeUrl` field to generate a QR image.

## In Mobile Device Browsers

    POST /wxpay/mobile/{standard|premium}/{year|month}

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```

### Input

None.

### Output

```json
{
    "ftcOrderId": "FT8AD02A2F3F0E44D2",
    "listPrice": 258,
    "netPrice": 0.01,
    "appId": "***REMOVED***",
    "mWebUrl": "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=wx031227207396425c62fb29a30807789694&package=865496058"
}
```

Client should redirect user to `mWebUrl`.

## In Wechat Browser

    POST /wxpay/jsapi/{standard|premium}/{year|month}

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```
   
### Input

```json
{
  "openId": "xxxxxxx"
}
```

### Output

```json
{
    "ftcOrderId": "FT4E74602E64ABFB89",
    "listPrice": 258,
    "netPrice": 0.01,
    "appId": "***REMOVED***",
    "timestamp": "1554265974",
    "nonce": "91dd2e4b9a92f730ea95",
    "pkg": "Sign=WXPay",
    "signType": "MD5",
    "signature": "1EFFFCDF30CCB94B10AADD69615C222D"
}
```

NOTE: the field names are slightly different from wechat requirements since we do not want to follow wechat's irregular spelling:

```
appId: appId
timeStamp: timestamp
nonceStr: nonce
package: pkg
signType: signType
paySign: signature
```

## Native App

    POST /wxpay/app/{standard|premium}/{year|month}

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: android | ios
X-Client-Version: 0.3.0
```

### Input 

None

### Response

```json
{
    "ftcOrderId": "FT89EFC6849EBEEEB3",
    "listPrice": 258,
    "netPrice": 0.01,
    "appId": "***REMOVED***",
    "partnerId": "1504993271",
    "prepayId": "wx0314214674083479994c12e90116519583",
    "timestamp": "1554272506",
    "nonce": "cf5ce8bf6321af407ade",
    "pkg": "Sign=WXPay",
    "signature": "EDDB1B107534AC06950F3C7343B28421"
}
```

App should then call wechat SDK:

```kotlin
wxApi = WXAPIFactory.createWXAPI(context, "wechat-app-id")
wxApi.registerApp("wechat-app-id")

val req = PayReq()
req.appId = wxOrder.appid
req.partnerId = wxOrder.partnerId
req.prepayId = wxOrder.prepayId
req.nonceStr = wxOrder.nonce
req.timeStamp = wxOrder.timestamp
req.packageValue = wxOrder.pkg
req.sign = wxOrder.signature

val result = wxApi.sendReq(req)
```

## Wxpay Unified Order [Deprecate]

    POST /wxpay/unified-order/{tier}/{cycle}

Request server to create wxpay's unified order. This is used only by Android app with version under 2.0.4.

### Input

`tier` must be one of `standard` or `premium`.

`cycle` must be one of `year` or `month`.

Request header must contain `X-User-Id: <uuid>` if user logged in with FTC account or `X-Union-Id: <wechat union id>` if used logged in with Wechat OAuth. If an FTC account is already bound to a Wechat account, and user logged in via Wechat OAuth, you **MUST** always use `X-User-Id: <uuid>` and should never use `X-Union-Id`.

Request header should also contain:
```
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
    "message": "Already a subscribed user and not within allowed renewal period."
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
    "ftcOrderId": "string",
    "listPrice": 258,
    "netPrice": 258,
    "appid": "wx app id",
    "partnerid": "wx mch id",
    "prepayid": "wx created prepay id",
    "package": "Sign=WXPay",
    "noncestr": "string",
    "timestamp": "unix timestamp",
    "sign": "string"
}
```

## Query Order

    GET /wxpay/query/{orderId}
    
 Query the payment result of an order. `orderId` is the the value included in the above responses.
 
 ### Headers
 
```
X-User-Id
X-Union-Id
X-App-Id
```

`X-App-Id` is wechat app id that is used to create this order. Client is required to send it to API due to the fact that we used multiple wechat apps to create orders. If client does not tell which app id is used to create the original order, we have no way to sign the query parameters correctly.

### Response

* `400 Bad Request`

If `orderId` is not provided;

* `500 Internal Server Error`

If query against wechat server failed.

* `404 Not Found`

If wechat response contains `ORDERNOTEXIST`.

* `200 OK`

```json
{
    "paymentState": "SUCCESS | REFUND | NOTPAY | CLOSED, REVOKED | USERPAYING | PAYERROR",
    "paymentStateDesc": "支付成功",
    "totalFee": 1,
    "transactionId": "4200000252201903069440709666",
    "ftcOrderId": "FT1D3CEDDB2599EFB9",
    "paidAt": "2019-03-06T07:21:18Z"
}
```