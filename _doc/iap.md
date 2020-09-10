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

### Paths

* POST `/apple/verify-receipt`
* POST `/apple/subscription`
* PATCH `/apple/subscription/<original_transaction_id>`
* POST `/apple/link`
* DELETE `/apple/link`
* GET `/apple/receipt/<original_transaction_id>`
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

* If any error occurred while sending request to Apple's verification request, like network error, timeout, etc.. This does not indicates the receipt is invalid.

#### `422 Unprocessable`

* If `receiptData` field is missing or empty.

```json
{
  "message": "receipt-data missing",
  "error": {
    "field": "receiptData",
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

#### `200 OK`

Apple's response is transferred to client as is. See https://developer.apple.com/documentation/appstorereceipts/responsebody.

Please note there is an error in Apple's doc saying the type of`latest_receipt` is `byte`. It is a string type, a string representation of the underlying byte array. You can safely ignore this field.

## Create/Update a Subscription

```
POST /apple/subscription
```

This is similar to the above endpoint, with response body as the only difference. The reponse contains only the essential information to identify a subscrition.

## Link IAP

```
POST /apple/link
```

This links apple subscription to an FTC account.

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "the original transaction id as acquired from the last endpoint."
}
```

### Response

#### `400 Bad Request` same as receipt verification

#### `422 Unprocessable`

In addition to the `422` in the above Verify Receipt section, there are other possible unprocessable errors:

* If this IAP is already linked to another FTC account, indicating user is trying to link one IAP to multiple FTC accounts. This is a possible cheat.

```
      Apple ID A
     /         \ <- This is not allowed
FTC ID A      FTC ID B
```

```json
{
  "message": "one apple subscription cannot be linked to multiple FTC account",
  "error": {
    "field": "iap_membership",
    "code": "already_linked"
  }
}
```

* If the linking target of FTC account is already linked to another IAP, indicating user might switch Apple ID:

```
Apple ID A      Apple ID B
     \           /
      \        / <- Requires severing link to Apple ID A
       FTC ID A
```

```json
{
  "message": "target ftc account is already linked to another apple subscription",
  "error": {
    "field": "ftc_membership",
    "code": "already_linked"
  }
}
```

* If the linking target already has a valid membership purchased via other channels, like Alipay, Wechat Pay, or Stripe:

```json
{
  "message": "target ftc account already has a valid membership",
  "error": {
    "field": "ftc_membership",
    "code": "valid_non_iap"
  }
}
```

#### `200 OK`

When a linking request comes in, the API will first try to retrieve membership data by both apple's original transaction id and ftc's uuid (or wechat's union id for wechat-only login), and both sides current status will be checked.

Linking is only allowed if two sides meet the following conditions (zero value here means it is not found in database):

* Both sides are zero values;
* Both sides are not zero values but they are equal, meaning both sides' uuid and original transaction id are equal;
* IAP side is not zero, indicating this IAP is already linked to an FTC account. Since they are not equal, the requested FTC account must be a new one. We can deduce that user is trying to link the same Apple ID to a new FTC account. This is a possible cheat.
* The only possibility left is IAP is zero while FTC is non-zero. If FTC membership is not valid (not existing or expired), we should allow user to link to it. FTC's previous membership data will be overridden. Otherwise the linking the denied.

Response body is an instance `Membership`

```json
{
    "id": "mmb_TP3gFqbuXRvD",
    "ftcId": "748bc0c8-f778-4616-8dff-7f4a8f4dd411",
    "unionId": null,
    "tier": "standard",
    "cycle": "month",
    "expireDate": "2019-11-22",
    "payMethod": "apple",
    "autoRenew": false,
    "status": null
}
```

* `id: string` a unique index id in DB.
* `ftcId: string | null` FTC's uuid
* `unionId: string | null` Wechat's union id. `ftcId` and `unionId` won't both be `null`. At least one of them will be available.
* `tier: standard | premium` An enum for the subscription tier.
* `cycle: year | month` An enum for billing cycle.
* `expireDate: string`
* `payMethod: apple | alipay | wechat | stripe` An enum specifying from which payment channel the current membership is purchased.
* `autoRenew: boolean | null` Only available for Stripe of Apple IAP.
* `status: active | canceled | incomplete | incomplete_expired | past_due | trialing | unpaid | null` An enum of Stripe's subscription and billing status.

There will be no `204 No Content` response for account linking.

### Unlinking

```
DELETE /apple/link
```

If user switched Apple ID, we should allow user to link the new Apple ID to previous FTC account which is already linked to its old Apple ID. To do this, client should ask user to manually unlink this FTC account from the old Apple ID.

*Unlinking* here is actually a delete operation.

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "orignal transaction id"
}
```

### Response

`400 Bad Request` and `422 Unprocessable` are identical to *Verify Receipt* section.

`204 No Content` if deleted successfully.

This is an idempotence operation. You always get `204` no matter how many times the same user sends request to this endpoint.

## Get a Receipt

```
GET /apple/receipt/<original_transaction_id>
```

## WebHook

```
POST /webhook/apple
```

Handles Apple's server-to-server notification.
