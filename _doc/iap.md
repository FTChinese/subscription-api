## Apple IAP

### Verify Receipt and Optionally Link Account

```
POST /apple/verify-receipt
```

#### Headers

```
X-User-Id: <ftc-uuid>
X-Union-Id: <wechat-union-id>
```

These two headers are optional. If absent, it only performs receipt verification; otherwise it will try to link the IAP to FTC account as specified by the header.

If user logged in with email, use the FTC uuid as the value of `X-User-Id`.

If user logged in with Wechat, pass Wechat's union id to `X-Union-Id`.

If email is linked to wechat, pass both of the two ids.

#### Request Body

```json
{
  "receipt-data": "the base64 encode apple receipt"
}
```

#### Response

##### `400 Bad Request` 

* If request body cannot be parsed as valid JSON;

* If any error occurred while sending request to Apple's verification request, like network error, timeout, etc.. This does not indicates the receipt is invalid.

##### `422 Unprocessable` 

* If `receipt-data` field is missing or empty.

```json
{
  "message": "receipt-data missing",
  "error": {
    "field": "receipt-data",
    "code": "missing_field"
  }
}
```

* If the verification response is not valid: `status` is not `0`; bundle id does not match; `latest_receipt_info` field is empty.

```json
{
  "message": "verification response is not valid",
  "error": {
    "field": "receipt-data",
    "code": "invalid"
  }
}
```

##### `204 No Content`

If neither `X-User-Id` nor `X-Union-Id` is passed and receipt data is valid.

It only indicates the receipt data itself is valid. It does not convey anything about the subscription status.

##### `200 OK`

If this is a link request and user account is linked.

Response body is an instance `Membership`

```json
{
  "id": "a unique id in db",
  "ftcId": "string | null",
  "unionId": "string | null",
  "tier": "standard | premium",
  "cycle": "year | month",
  "expireDate": "2020-12-12",
  "payMethod": "apple | alipay | wechat | stripe",
  "autoRenewal": true,
  "status": "active | canceled | incomplete | incomplete_expired | past_due | trialing | unpaid | null"
}
```