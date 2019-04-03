# Notification

Those endpoints are used to receive payment result from Alipay and Wechat pay, not being used by us. **DO NOT** impose any access rights on them.

## Wechat Pay Notification

    POST /callback/wxpay

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
    "appid":"***REMOVED***",
    "openid":"ob7fA0h69OO0sTLyQQpYc55iF_P0",
    "trade_type":"APP"
}
```

## Alipay Notification

    POST /callback/alipay