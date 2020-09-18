# Apple IAP

All request must have access token set in its header using OAuth 2.0 Bearer Token, not including the webhook endpoints which are not used by us:

```
Authorization: Bearer <your-access-token>
```

See https://tools.ietf.org/html/rfc6750#section-2.1

Generate an access token for your app in Superyard. *DO NOT* use your personal access tokens here since personal access tokens might be revoked any moment.

## Conventions

Client should take HTTP status above 400 as error. Error response always returned a body:

```json
{
  "message": "a short human readable message in English",
  "error": {
    "field": "where_error_happened",
    "code": "error_reason"
  }
}
```

`message` is always present.

`error` is only present for `422` response. You can use the combination of `error.field` and `error.code` to determine what goes wrong.

## Base URL

* Sandbox: `https://www.ftacademy.cn/api/sandbox`
* Production: `https://www.ftacademy.cn/api/v1`

### Difference between sandbox and production mode

The sandbox and production URLs are running the same code base, manipulate the same db and tables. Their difference mostly lies in setting webhook URLs of Alipay, Wxpay, Apple and Stripe, as well as Apple's receipt verification endpint..

### Test Account

When you logged in with a test account, which can be found in Superyard, requests should go to sandbox API; otherwise use production API.

For Alipay and Wxpay, subscription prices will be set to 0.01 so that you don't need to be bothered with refunding.

### Paths

* POST `/apple/verify-receipt` Verify a receipt.
* POST `/apple/link` Link an IAP to an FTC account.
* POST `/apple/unlink` Unlink IAP from an FTC account.
* POST `/apple/subs` Verify a receipt and get the essential subscription data.
* GET `/apple/subs?page=<int>&per_page=<int>` Load a list of existing IAP subscription.
* GET `/apple/subs//<original_transaction_id>` Load a single IAP subscription.
* PATCH `/apple/subs/<original_transaction_id>` Refresh an existing IAP subscription against Apple verification.
* GET `/apple/receipt/<original_transaction_id>` Load a single IAP subscription together with the receipt file.
* POST `/webhook/apple`

## Verify Receipt

```
POST /apple/verify-receipt
```

### Request Body

```json
{
  "receiptData": "the base64 encode apple receipt"
}
```

### Response

#### `400 Bad Request`

* If request body cannot be parsed as valid JSON;

#### `422 Unprocessable`

* If request body has invalid field, e.g, `receiptData` field is missing or empty.

```json
{
  "message": "Missing required field",
  "error": {
    "field": "receiptData",
    "code": "missing_field"
  }
}
```

* If the verification response is not valid, those fields in the response will be checked: 

`status` is not `0`: 

```json
{
  "message": "The data in the receipt-data property was malformed or missing",
  "error": {
    "field": "status",
    "code": "invalid"
  }
}
```

`bundle_id` does not match:

```json
{
    "message": "The data in the receipt-data property was malformed or missing",
    "error": {
        "field": "bundle_id",
        "code": "invalid"
    }
}
```

`latest_receipt_info` field is empty.

```json
{
    "message": "Latest receipt info should not be empty",
    "error": {
        "field": "latest_receipt_info",
        "code": "missing_field"
    }
}
```

#### `500 Interval Server Error`

* This indicates an error occurred when building http request which will be sent to Apple. It does not mean Apple responded an error.

* Apple responded with something, but the response body cannot be parsed as valid JSON.

#### `200 OK`

Apple's response is transferred to client as is. See https://developer.apple.com/documentation/appstorereceipts/responsebody.

Please note there is an error in Apple's doc saying the type of`latest_receipt` is `byte`. It is a string type, a string representation of the underlying byte array. You can safely ignore this field.

## Link IAP

```
POST /apple/link
```

Links apple subscription to an FTC account. This endpoint does not perform receipt validation. It only checks if an `originalTxId` is found in the current database, and tries to link it to the FTC account if found.

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "the original transaction id"
}
```

### Response

#### `400 Bad Reqeust`

* If request body cannot be parsed as JSON;

#### `422 Unprocessable`

* If request body is not valid. See `/apple/verify-receipt` part.

* If IAP side already has a membership, it means IAP is linked to another FTC account:

```json
{
  "message": "An apple subscription cannot link to multiple FTC accounts",
  "field": "iap_membership",
  "code": "already_linked"
}
```

* If FTC side is linked to an IAP:

```json
{
  "message": "Target ftc account is already linked to another apple subscription",
  "field":   "ftc_membership",  
  "code":    "already_linked"
}
```

* If FTC side exists and not-expired, and not-iap:

```json
{
  "message": "Target ftc account already has a valid non-iap membership",
  "field": "ftc_membership",
  "code": "valid_non_iap"
}
```

#### `400 Not Found`

* If `ftcId` is not found;
* If Apple subscription is not found;

#### `200 OK`

Response body is an instance of `Membership`

```json
{
    "ftcId": "748bc0c8-f778-4616-8dff-7f4a8f4dd411",
    "unionId": null,
    "tier": "standard | premium",
    "cycle": "month | year",
    "expireDate": "2019-11-22",
    "payMethod": "apple",
    "ftcPlanId": null,
    "stripeSubsId": null,
    "autoRenew": false,
    "status": null,
    "appleSubsId": "apple original transaction id",
    "b2bLicenceId": null
}
```

### Link Strategy

When performing linking, we need to take into account current memberships of both FTC side and IAP side. There's a lot of combinations and only a few of them are allowed, as shown by the following table:


| FTC\IAP     | None   | Not-Expired | Expired |
| ----------- | ------ | ----------- | ------- |
| None        |  Y     |      N      |  N      |
| Not-Expired |  N     |      N      |  N      |
| Expired     |  Y     |      N      |  N      |

For IAP side, it must not have a membership prior to linking; otherwise it indicates IAP is already linked to another FTC and user is  trying to link the same IAP to multiple FTC accounts. Therefore, only the first column has cases to be allowed. Among them, if the FTC side has a non-expired membership as in column 1, row 2, it is not allowed to link since this will cause data overriding. 

## Unlinking

```
POST /apple/link
```

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "the original transaction id"
}
```

*Unlinking* here is actually a delete operation.

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "original transaction id"
}
```

### Response

* `400 Bad Request`

* `422 Unprocessable`

* `204 No Content` if unlinked successfully.

## Verify Receipt and Get Subscription

```
POST /apple/subs
```

This is almost the same operation as performed by `/apple/verify-receipt`, except the response is a condensed version of Apple's parsed receipt. We call it Apple's `Subscription`.

See `/apple/verify-receipt` for request.

If succeeded, the response body will be an instance of `apple.Subscription`:

```json
{
  "environment": "Production | Sandbox",
  "originalTransactionId": "30000781417036",
  "lastTransactionId": "30000781417036",
  "productId": "com.ft.ftchinese.mobile.subscription.member.monthly",
  "purchaseDateUtc": "2020-06-11T02:53:00Z",
  "expiresDateUtc": "2020-07-11T02:53:00Z",
  "tier": "standard",
  "cycle": "month",
  "autoRenewal": true,
  "createdUtc": "2020-06-11T02:56:12Z",
  "updatedUtc": "2020-06-11T02:56:12Z"
}
``` 

## List Subscription

```
GET /apple/subs?page=<int>&per_page<int>
```

Get a list of Subscription. Query parameter `page` specifies the current page, and `per_page` specifies how many items should be retrieved per page.

### Response

```json
{
  "total": 20,
  "page": 1,
  "limit": 20,
  "data": [
    {
        "environment": "Sandbox",
        "originalTransactionId": "1000000619244062",
        "lastTransactionId": "1000000619244062",
        "productId": "com.ft.ftchinese.mobile.subscription.member.monthly",
        "purchaseDateUtc": "2020-01-25T00:19:53Z",
        "expiresDateUtc": "2020-01-25T00:24:53Z",
        "tier": "standard",
        "cycle": "month",
        "autoRenewal": false,
        "createdUtc": "2020-09-15T04:04:16Z",
        "updatedUtc": "2020-09-18T01:20:06Z"
    }
  ]
}
```

The `data` array is sorted in descending order by `updatedutc`.

## Load a Single Subscription

```
GET /apple/subs/<original_transaction_id>
```

Retrieve a single `apple.Subscription` by original transaction id.

## Refresh a Subscription

```
PATCH /apple/subs/<original_transaction_id>
```

This allows you to refresh an existing Apple Subscription based on the receipt previously saved. If client does not know what the original transaction id is, post device's receipt to the above said `POST /apple/subs` to get it.

Response is an instance of `apple.Subscription` extracted from the verified receipt.

Workflow of the refreshing process:

1. Use the original transaction id to find the `Subscription` from db;
2. Build file name from the `originalTransactionId` and `envrionment` field. For example, `1000000322563042_Sandbox`;
3. Use the file name to read the receipt from disk;
4. Verify the receipt as in the `/apple/verify-receipt`, and then the receipt file. The `Subscription` for this original transaction id is updated and returned in response.

## Get a Receipt

```
GET /apple/receipt/<original_transaction_id>
```

This is almost the same as `/apple/subs/<orginial_transaction_id>` with an additional field:

```json
{
    "environment": "Production",
    "originalTransactionId": "30000781417036",
    "lastTransactionId": "30000781417036",
    "productId": "com.ft.ftchinese.mobile.subscription.member.monthly",
    "purchaseDateUtc": "2020-06-11T02:53:00Z",
    "expiresDateUtc": "2020-07-11T02:53:00Z",
    "tier": "standard",
    "cycle": "month",
    "autoRenewal": true,
    "createdUtc": "2020-06-11T02:56:12Z",
    "updatedUtc": "2020-06-11T02:56:12Z",
    "receipt": "receipt file"
}
```

## WebHook

```
POST /webhook/apple
```

Handles Apple's server-to-server notification.
