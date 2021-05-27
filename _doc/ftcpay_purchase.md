# Alipay

`X-Client-Type` is `web` for both desktop and mobile browsers, `android` for Android app annd `ios` for iOS app.

## Error Response

Those endpoints share similar error response:

* `401 Unauthorized` if request header does not contain `X-User-Id` or `X-Union-Id`.

* `400 Bad Request`

If the combination of `tier` and `cycle` is not one of `standard_month`, `standard_year` or `premium_year`


## In Desktop Browser

    POST /alipay/desktop/{standard|premium}/{year|month}<?return_url=xxxx>

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```

The query parameter `return_url` is optional. It specifies to where user should be redirected after payment. If omitted, it defaults to `http://www.ftacademy.cn/api/v1/redirect/alipay/done` (see `PayRouter.aliReturnURL()`), which will perform a redirection again to `http://next.ftchinese.com/user/subscription/alipay/callback`, which will again redirect to `http://next.ftchinese.com/user/subscription` if everything works fine.

Such complex redirection strategy is used to circumvent Alipay's restriction that the redirected-to url must be under the registered domain in Alipay.

### Input

None

### Response

`200 OK`:

```json
{
    "ftcOrderId": "FT16561420686FD082",
    "listPrice": 258,
    "netPrice": 0.01,
    "payUrl": "https://openapi.alipay.com/gateway.do?app_id=2018053060263354&biz_content=%7B%22subject%22%3A%22FT%E4%B8%AD%E6%96%87%E7%BD%91+-+%E5%B9%B4%E5%BA%A6%E6%A0%87%E5%87%86%E4%BC%9A%E5%91%98%22%2C%22out_trade_no%22%3A%22FT16561420686FD082%22%2C%22total_amount%22%3A%220.01%22%2C%22product_code%22%3A%22FAST_INSTANT_TRADE_PAY%22%2C%22goods_type%22%3A%220%22%7D&charset=utf-8&format=JSON&method=alipay.trade.page.pay&notify_url=http%3A%2F%2Fwww.ftacademy.cn%2Fapi%2Fsandbox%2Fcallback%2Falipay&return_url=http%3A%2F%2Fwww.ftacademy.cn%2Fapi%2Fsandbox%2Fredirect%2Falipay%2Fnext-user&sign=JDmH2%2B8yPdPvo9gjavN8mhi%2FqfiBYlL2GZRXXStYALuWA9vNtKC58vGD8CfQfO6JQ%2FaRKT7kbtCsJAX9FnN6MKIzwCa4yO9SrfsBWQpXKOEXc9vX%2Bi5h9Xi4YgsCXvPSfukBw%2F47qi23iTFX2FXBD%2BKHPHHmJpWna3q%2BwHVeWsAMcc4eR3WkjxrKGfNFNA%2F%2B2P55DmTbDi30zP5C7I%2Bz5bJijMcFCXL%2Bd8MarQN3v5aIRFfZBJxDDYjPdneQZlu0F8LYnA5lquBitqzc1WvJIO9MjI7jlOM0AkgOTYtRyIIRdJ1Eyns7sj8LEyFvK59lL%2BxB5rHpt9jetJSr3j7VJQ%3D%3D&sign_type=RSA2&timestamp=2019-04-03+15%3A11%3A32&version=1.0"
}
```

Client should redirect user to the `payUrl`.

## In Mobile Browser
    
    POST /alipay/mobile/{standard|premium}/{year|month}<?return_url=xxx>

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```

### Response

`200 OK`

```json
{
    "ftcOrderId": "FT582F909C1970147D",
    "listPrice": 258,
    "netPrice": 0.01,
    "payUrl": "https://mclient.alipay.com/cashier/mobilepay.htm?alipay_exterface_invoke_assign_target=invoke_8b8d418f023c52c32d5cfae15cf88567&alipay_exterface_invoke_assign_sign=_hw9oe_ww_b3ofj_j_a_ibr3y_k_y%2F5gwe_m_h8kol38ky_q_r_d_m_tt_bp_x_d_nry_ri_y_aw%3D%3D"
}
```

Client should redirect to the `payUrl`, which will call app Zhifubao.

User will be redirect back to `return_url`. See above explanation.

## Native App

    POST /alipay/app/{standard|premium}/{year|month}

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: android | ios
X-Client-Version: 0.3.0
```

### Response

`200 OK`:

```json
{
    "ftcOrderId": "FT981F8403B38C2937",
    "price": 258,
    "listPrice": 258,
    "netPrice": 0.01,
    "param": "app_id=2018053060263354&biz_content=%7B%22subject%22%3A%22FT%E4%B8%AD%E6%96%87%E7%BD%91+-+%E5%B9%B4%E5%BA%A6%E6%A0%87%E5%87%86%E4%BC%9A%E5%91%98%22%2C%22out_trade_no%22%3A%22FT981F8403B38C2937%22%2C%22total_amount%22%3A%220.01%22%2C%22product_code%22%3A%22QUICK_MSECURITY_PAY%22%2C%22goods_type%22%3A%220%22%7D&charset=utf-8&format=JSON&method=alipay.trade.app.pay&notify_url=http%3A%2F%2Fwww.ftacademy.cn%2Fapi%2Fsandbox%2Fcallback%2Falipay&sign=e69TE%2F0GJWeHkOKFJbWYqD%2FrR0SM0WzJtbCgJx0tc4NkCFV0K3Z%2F2cHA%2B85qeIds1PV0R%2BxCOwYZAgQmB2G0xDRtSRfN9PueSmbwF9m8z5MCTlkbKJgoK0BlaHemZgOdtqYjz%2B%2Bl5lha3keanvPOl6wAMNhRevxVRBaFCDlfvvz%2B1cNTnbBBWKzKrLUF2Ctaao7LXOQaLOMhlGUB8BCi2nyBy9C4IbxQb1nEUvkp4L3up11tveJuVYyZ22bhM1LzxUzD45%2FcL%2BdRbxXyFSWhzDVEeNVj4HVP4H9ai8GP9ixOQD6nT87abVUNuLEffCAolpVnMW7T5lpM%2Fx1m1SA9Pg%3D%3D&sign_type=RSA2&timestamp=2019-04-03+15%3A14%3A40&version=1.0"
}
```

Client should then perform:

```kotlin
val payResult = withContext(Dispatchers.IO) {
    PayTask(activity).payV2(param, true)
}

val resultStatus = payResult["resultStatus"]
if (resultStatus != "9000") {
    // failed
}
```

## Create Alipay Order [Deprecate]

    POST /alipay/app-order/{tier}/{cycle}

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

If sign request parameters failed;

* `403 Forbidden`

If this user is already a member and current date is not within the allowed renewal period.

```json
{
    "message": "Already a subscribed user and not within allowed renewal period."
}
```

* `404 Not Found` if current does not exist.

* `200 OK`

```json
{
    "ftcOrderId": "string",
    "netPrice": 258,
    "listPrice": 258,
    "param": "string"
}
```

## Alipay Redirect URL

    GET /redirect/alipay/next-user
 
 Example redirect query parameters:
 
 ```
 /redirect/alipay/next-user?charset=utf-8&out_trade_no=FTD94BB6AB13EE1DA3&method=alipay.trade.page.pay.return&total_amount=0.01&sign=P1acUcD4jMduPDI6Qn%2FJVAinAmxJlMz%2BdAiIrfBQnUJXSzsm4gFZtpwaPok2ar9Gg7imjkaTP2FpqqN0ISk3LaTbU%2BVS%2BhI%2B2yRBQylQRwnDexV9dMD848y8PF%2BQji7Qr3e5qXiXHgG4E%2B6VzNewHyTGKuDlEkXTtQULbqyOhCv3HmU%2FFopemb7JQ3C9BtA%2BsHPoZ68jxkTNvtyIf3Fi8iFTXe9rsuAzZStxBbAxBxXD2TxuxReAO6roCCxjeFiC7HsYhWPIRgia9atH9gS3LmyZ8szy7c3rn14c9QV13MXyTKEA9t8j4Lhydn%2Bs0XOnVVqJjZlBuRfxLZndtl2Yww%3D%3D&trade_no=2019040222001440031024509039&auth_app_id=2018053060263354&version=1.0&app_id=2018053060263354&sign_type=RSA2&seller_id=2088521304936335&timestamp=2019-04-02+17%3A08%3A14
 ```
 
 Those fields are:
 
 * `app_id`
 * `auth_app_id`
 * `charset=utf-8`
 * `method=alipay.trade.page.pay.return`
 * `out_trade_no=FT99a55d609736c4fa`
 * `seller_id`
 * `sign`
 * `sign_type=RSA2`
 * `timestamp=2019-03-29 14:06:58`
 * `total_amount=0.01`
 * `trade_no=2019032922001440031023691208`
 * `version=1.0`
 
 Here the API only verifies whether the signature is correct. If signature is invalid, redirect-to url will receive query parameters:
 
 ```
 error=invalid_signature
 error_descripiton=something-went-wrong
 ```
 
 If signature is correct, then the original query parameters will appended to `http://next.ftchinese.com/user/subscription/alipay/callback`.

# Wechat Pay

## Shared Input and Output

### Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: web
X-Client-Version: 0.3.0
X-User-Ip: 248.250.115.148
X-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36
```

`X-User-Id` is FTC account's uuid if user logged in with FTC account while `X-Union-Id` is wechat's union id if user logged in via Wechat OAuth. You should provide at least one of them so that user could be identified. If user' FTC account is linked to Wechat account, you should provide both of them.

Since this is used only by desktop browsers, the API has no way to know user's ip and user agent. Thus they must be forwarded by the client app.

Native apps do not needs provide `X-User-Ip` and `X-User-Agent` since the request is issued directly from user's device.

### Error Response

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

## Native App

```
POST /wxpay/app
```

### Required Headers

```
X-User-Id: 48cdc7d4-a7ea-4b3e-9638-d423975810f4
X-Union-Id: tvSxA7L6cgl8nwkrScm_yRzZoVTy
X-Client-Type: android | ios
X-Client-Version: 0.3.0
```

### Input

```typescript
interface WxPayInput {
    tier: 'standard' | 'premium';
    cycle: 'year' | 'month';
    planId: string;
}
```

### Workflow



### Client Implementation

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

## In Desktop Browser

    POST /wxpay/desktop/{standard|premium}/{year|month}

This is used to create a QR code in desktop browsers. It corresponds to `trade_type=NATIVE` as specified by wechat. See the official documentation [扫码支付](https://pay.weixin.qq.com/wiki/doc/api/native.php?chapter=6_1)


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
    "appId": "wxacddf1c20516eb69",
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
    "appId": "wxc1bc20ee7478536a",
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
    "appId": "wxacddf1c20516eb69",
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
