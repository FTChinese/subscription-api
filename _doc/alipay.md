# Alipay

## Create Alipay Order

    POST /alipay/app-order/{tier}/{cycle}

### Input

Exactly the same as Wxpay Unified Order.

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
    "message": "Already a subscribed user and not within allowed renewal period.",
}
```

* `404 Not Found` if current does not exist.

* `200 OK`

```json
{
    "ftcOrderId": "string",
    "param": "string" // Pass this string to alipay sdk.
}
```

## Verify Ali App Pay Result

    POST /alipay/verify/app-pay

### Input

In your app, when you called Zhifubao, if will show a popup window on top of your app. After you confirmed payment and the popup window goes away, you app will get a map:
```json
{
    "resultStatus":"9000", 
    "result":"", 
    "memo": ""
}
```

Note `result` is a string, not a map. Post the string directly to this endpoint:
```
{"alipay_trade_app_pay_response":{"code":"10000","msg":"Success","app_id":"","auth_app_id":"","charset":"utf-8","timestamp":"2018-10-28 16:49:31","out_trade_no":"FT0055301540716534","total_amount":"0.01","trade_no":"2018102822001439881007782559","seller_id":"2088521304936335"},"sign":"MHrLSKA3KUKxsN9Yuhnzqbj5jpnSQ8drar5nt3gQJ0OzSTmmaYvYhEPEf/Qf6T+3t4UAnWmbRRGuHqruDK2/AuH+xtmhElPFLXo9dnkduUe5c15/AKtW6V2SWs+TGmSi38Wb/3NgeINtlSSxGnLXsW3uzbnybEd0E/L4nyqaKZ+yF3GWsWAsLzgf/O9y5ntpc7st3Vu1I2icipp34N9a4UbnOML0/kPuLls09K6/w461AAXh2GE4+L103lp/M4QFd5Lghauod75VctKI/xro06jIEjRkojFOOry+dugqEDxUQX+3CHzqOojub6ozD5GTZUV0ynOZCQA4iX+oOZ52lw==","sign_type":"RSA2"}
```

You need to manually extract this JSON-string to get the value of `alipay_trade_app_pay_response` -- an extremely stupid design.

If you parse it into a map and then verify the pased data structure, you'll never pass -- JSON is unordered and order is important in digital signature.

### Response

* `400 Bad Request` if parsing JSON failed.

* `404 Not Found` if the order is not found.

* `422 Unprocessable Entity`

if signature is not valid:
```json
{
    "message": "",
    "error": {
        "field": "sign",
        "code": "invalid"
    }
}
```

if signature is valid but not correct:
```json
{
    "message": "",
    "error": {
        "field": "sign",
        "code": "incorrect"
    }
}
```

if `app_id` is wrong:
```json
{
    "message": "",
    "error": {
        "field": "app_id",
        "code": "incorrect"
    }
}
```
if `total_amount` does not match the one recorded in database.
```json
{
    "message": "",
    "error": {
        "field": "total_amount",
        "code": "incorrect"
    }
}
```

* `200 OK`
```json
{
    "code": "",
    ""
}
```

## Notification