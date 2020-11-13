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

## Paths

* POST `/apple/verify-receipt` Verify a receipt.
* POST `/apple/link` Link an IAP to an FTC account.
* POST `/apple/unlink` Unlink IAP from an FTC account.
* POST `/apple/subs` Verify a receipt and get the essential subscription data.
* GET `/apple/subs?page=<int>&per_page=<int>` Load a list of existing IAP subscription.
* GET `/apple/subs//<original_transaction_id>` Load a single IAP subscription.
* PATCH `/apple/subs/<original_transaction_id>` Refresh an existing IAP subscription against Apple verification.
* GET `/apple/receipt/<original_transaction_id>` Load a single IAP subscription together with the receipt file.
* POST `/webhook/apple`

## The verification process

The verification process is performed by multiple endpoints:

* POST `/apple/verify-receipt`;
* POST `/apple/subs`;
* PATCH `/apple/subs`.

We dedicate this section to its details.

1. Send http request to App Store endpoint as required by Apple. If any errors occurred, we send `500 Internal Error` to client.

2. Then we validate various fields in the response to ensure it is valid. If Apple's response is not valid, we will send the following errors to our clients:

If `status` is not `0`, it indicates error: 

```json
{
  "message": "The data in the receipt-data property was malformed or missing",
  "error": {
    "field": "status",
    "code": "invalid"
  }
}
```

If `bundle_id` does not match:

```json
{
    "message": "The data in the receipt-data property was malformed or missing",
    "error": {
        "field": "bundle_id",
        "code": "invalid"
    }
}
```

If `latest_receipt_info` field is empty.

```json
{
    "message": "Latest receipt info should not be empty",
    "error": {
        "field": "latest_receipt_info",
        "code": "missing_field"
    }
}
```

3. Parse the response to get user's valid subscription.

4. Dissect the response so that we could save its various fields to relational database. The response has the following structure:

```json
{
  "environment": "sandbox or production",
  "latest_receipt": "the latest base64-encoded app receipt",
  "latest_receipt_info": [],
  "pending_renewal_info": [],
  "status": 0,
  "is-retryable": false,
  "receipt": {}
}
```

The data under `latest_receipt`, `latest_receipt_info`, `pending_renewal_info` and `receipt` are saved into those tables:

| Field                | Table Name                         |
| -------------------- | ---------------------------------- |
| receipt              | premium.apple_verification_session |
| latest_receipt_info  | premium.apple_transaction          |
| pending_renewal_info | premium.apple_pending_renewal      |
| latest_receipt       | file_store.apple_receipt           |

The `environment` field is saved along each row to all the tables mentioned here so that we know from which environment each row is generated.

Note the `latest_receipt` is original saved to disk. Later it is also save to Redis. Then we also save a copy to MySQL, therefore a receipt might appear in three places:

* Disk on ucloud;
* Redis
* MySQL

6. We create a data structure called `Subscripiton` which contains the essential data from the verification response and save it in `premium.apple_subscription`:

```json
{
  "environment": "",
  "originalTransactionId": "uniquely identify a apple subscription",
  "latestTransactionId": "the last transaction id",
  "productId": "",
  "purchaseDateUtc": "purchase date in iso format set in utc time zone",
  "expiresDateUtc": "expiration date",
  "tier": "standard | premium",
  "cycle": "month | year",
  "autoRenewal": true,
  "createdUtc": "when this row is inserted",
  "updatedUtc": "the last update time",
  "ftcUserId": "uuid if linked to ftc account; otherwise null"
}
```

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

### Workflow

1. Parse request body. Returns `400` if parsing failed

2. Validate request body. Returns `422` if `receiptData` missed:

```json
{
  "message": "Missing required field",
  "error": {
    "field": "receiptData",
    "code": "missing_field"
  }
}
```

3. Send receipt data to apple server. Returns `422` if any error occurred in the request to Apple, or response is not valid:

```json
{
  "message": "might be anything",
  "error": {
    "field": "verification",
    "code": "invalid"
  }
}
```

4. The response body is dissected and save into DB.

5. Create a `Subscription` base on the response and save it to `premium.apple_subscription`.

6. Apple's response is sent to client as is.

## Link IAP

```
POST /apple/link
```

Links apple subscription to an FTC account. This endpoint does not perform receipt validation. It only checks if an `originalTxId` is found in the current database, and tries to link it to the FTC account if found.

### Request Body

```json
{
  "ftcId": "xxxxx",
  "originalTxId": "xxxxx"
}
```

* `ftcId: string` FTC uuid. Required.
* `originalTxId: string` Apple subscription's original transaction id. Required.

### Link Policy

When performing linking, we need to take into account current memberships of both FTC side and IAP side. There's a lot of combinations and only a few of them are allowed, as shown by the following table:


| FTC\IAP     | None   | Not-Expired | Expired |
| ----------- | ------ | ----------- | ------- |
| None        |  Y     |      N      |  N      |
| Not-Expired |  N     |      N      |  N      |
| Expired     |  Y     |      N      |  N      |

For IAP side, it must not have a membership prior to linking; otherwise it indicates IAP is already linked to another FTC and user is  trying to link the same IAP to multiple FTC accounts. Therefore, only the first column has cases to be allowed. Among them, if the FTC side has a non-expired membership as in column 1, row 2, it is not allowed to link since this will cause data overriding. 

Case in Column 1, Row 2 has an edge case: The FTC side might be manually created from an IAP result by customer service. It has no payment method, not expired. Should we allow linking in such case? The answer is non-deterministic. Currently my approach is to compare the expiration date the FTC membership against IAP subscription. The IAP subscription expires later, allow linking; otherwise it is forbidden.

### Workflow

1. Parse request body. It must contain `ftcId` and `originalTxId`. If it cannot be parsed, respond `400 Bad Request`.

2. Validate request body. If any of the required field missed, respond `422 Unprocessable Entity` with body:

```json
{
  "message": "Missing required field",
  "error": {
    "field": "ftcId | originalTxId",
    "code": "missing_field"
  }
}
```

3. Retrieve user's account by `ftcId`. Respond `404` if account not found.

4. Retrieve IAP subscription by `originalTxId` from `premium.apple_subscription` table. Respond `404` if not found.

5. The retrieve subscription has a nullable field `ftcUserId`. If it is not null and not equal to the passed in `ftcId`, it indicates a new ftc account is trying to use an existing IAP which is already claimed by another ftc account. Since we should not allow multiple ftc accounts using the same IAP, this might be cheating. Respond `422` with:

```json
{
  "message": "Apple subscription is already claimed by another ftc account.",
  "error": {
    "field": "originalTxId",
    "code": "linked_to_other_ftc"
  }
}
```

This `ftcId` and `originalTxId` combination is also saved to `premium.apple_cheat` table so that we could examine them later.

6. Otherwise we set `ftcId` to the `ftc_user_id` column if it is still empty.

7. Now we have valid `FtcAccount` and iap `Subscription`, starting linking  them.

8. First retrieve `Membership` by the `ftc_vip.apple_subscription_id` column using the IAP's original transaction id. Then retrieve `Membershp` again by the `ftc_vip.vip_id` using the ftc id. Note if both are not found in DB, we do not treat them as error since zero value of `Membership` is a valid value to operate on.

9. Now that we have `FtcAccount`, existing iap side `Membership`, existing ftc side `Membership`, and an update-to-date iap `Subscription`, we'll validate whether the `Subscription` is allowed to link to `FtcAccount`.

10. If current iap side `Membership` is equal to ftc side `Membership`, it indicates 2 possibilities: 

    * Both sides are clean. You can go ahead. This is an initial link. An email will be sent after the operation finished.
    
    * Both sides retrieved the same membership, which means the account is already linked to this membership. The `Membership`'s expiration date will be synced to the `Subscription` if not equal.

11. The two `Membership`s are not equal. So at least one of them exists. If iap side `Membership` is not zero, it means iap already has a membership linked to an FTC account, and now another FTC account is trying to link to it. This is a possible cheating similar to step 5: **multiple ftc, single apple id**.

12. Since iap side has no membership, ftc must have one (otherwise it should fall into step 10). Now we should consider if the ftc side membership is also created from an iap (by checking the `app_subscription_id` column). This indicates the current user is already linked to an iap and now it is trying to link to a new iap. This is the case of a **single ftc, multiple apple id**. Such case might arise when a user switched Apple account and both Apple accounts have subscription. Respond `422`:

```json
{
  "message": "FTC account is already linked to another Apple subscription",
  "error": {
    "field": "ftcId",
    "code": "linked_to_other_iap"
  }
}
```

13. Now iap side has no membership while ftc side has membership. The above diagram narrows down to column 1. If ftc side membership is not expired (might purcahse via alipay or wxpay, or Stripe), we should not override it. Respond `422` unless step 14 happens:

```json
{
  "message": "FTC account already has a valid membership via non-Apple channel",
  "error": {
    "field": "ftcId",
    "code": "has_valid_non_iap"
  }
}
```

14. There's an edge case in step 13: ftc side `Membership` does not has `PaymentMethod` (`payment_method` column in db) field set, which might be true if it manually touched, we cannot determine the source of this membership. So we will check the ftc-side `Membership`'s expiration date against iap `Subscription` expiration date. If iap subscription expires later, override the ftc side.

15. Link done and send `200 OK`. Response body is an instance of updated `Membership`

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


## Unlink IAP

Sever links between an FTC account and Apple subscription.

```
POST /apple/unlink
```

### Request Body

```json
{
  "ftcId": "uuid",
  "originalTxId": "the original transaction id"
}
```

* `ftcId: string` Required
* `originalTxId: string` Required

### Workflow

1. Parse request body. Return `400` if parsing failed.

2. Validate request body. Save as step 2 in `Link IAP`.

3. Start a DB transaction. Retrieve membership by `originalTxId`. Return `404` if membership does not exist.

4. If `ftcId` is not equal to membership's `FtcID` field, returns `422` and db transaction rolled back:

```json
{
  "message": "IAP is not linked to the ftc account",
  "error": {
    "field": "ftcId",
    "code": "invalid"
  }
}
```

5. Delete this row from `premium.ftc_vip`. Roll back in case of db error.

6. Retrieve apple subscription by `originalTxId`. Roll back in case of db error.

7. If the subscription's `FtcUserID` field is equal to `ftcId`, set it to null.

8. Commit db transaction.

9. The deleted membership is archived into `premium.member_snapshot`.

10. The `ftcId` and `originalTxId` is save to `premium.apple_unlink_archive`.

11. An email is sent to user to notify user this unlink operation.

12. Returns `202 No Content` to indicate operation succeeded.

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
  "updatedUtc": "2020-06-11T02:56:12Z",
  "ftcUserId": ""
}
``` 

## List Subscription

Required header: `X-Ftc-Id: <uuid>`

```
GET /apple/subs?page=<int>&per_page<int>
```

Get a list of Subscription belong to a user. Query parameter `page` specifies the current page, and `per_page` specifies how many items should be retrieved per page.

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
        "updatedUtc": "2020-09-18T01:20:06Z",
        "ftcUserId": ""
    }
  ]
}
```

## Load a Single Subscription

```
GET /apple/subs/<original_transaction_id>
```

Retrieve a row from `premium.apple_subscription`.

## Refresh a Subscription

```
PATCH /apple/subs/<original_transaction_id>
```

This allows you to refresh an existing Apple Subscription based on the receipt previously saved. If client does not know what the original transaction id is, post device's receipt to the above said `POST /apple/subs` to get it.

Response is an instance of `apple.Subscription` extracted from the verified receipt.

### Workflow

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

## Accessing API from iOS

To access API you need to present an access token for each request, and the access token should be kept secret. Never leak it to public.

You can put the access token in Xcode configuration files and use `git-crypt` to encrypt it.

### XCode configuration file

In you Xcode project, click `File > New > File...`. Scroll to the `Other` section in the opened dialog. Select `Configuration Settings File`. You are asked to save the file. Name it `Keys.xcconfig`. Make sure nothing is check in the `Targets` section. Click `Create`.

Add the access token to it:

```
ACCESS_TOKEN = you_access_token
```

If you want to keep project files organized into folders, right-click on your project name in the navigator. Select `New Group` and give it a name `Keys`, for example.

Add another configuration file `Production.xcconfig` in the project root. Add the following content to it:

```
#include "Keys/Keys.xcconfig"
```

Open `Info.pllist` file of your project. Right-click and select `Add Row`. Give the key name `AccessToken` (or whatevery name you like) and set its value to `$(ACCESS_TOKEN)`.

To use the variable in Swift, create a new file `API.swift`, for example. Read the data from bundle:

```swift
struct Configuration {
    enum Error: Swift.Error {
        case missingKey, invalidValue
    }
    
    static func value<T>(for key: String) throws -> T where T: LosslessStringConvertible {
        guard let object = Bundle.main.object(forInfoDictionaryKey: key) else {
            throw Error.missingKey
        }
        
        switch object {
        case let value as T:
            return value
        case let string as String:
            guard let value = T(string) else { fallthrough }
            return value
        default:
            throw Error.invalidValue
        }
    }
}

struct API {
    static var accessToken: String {
        return try! Configuration.value(for: "AccessToken")
    }
}
```

Now you can use the variable from Xcode configuration files.

References:

* https://www.appcoda.com/xcconfig-guide/
* https://nshipster.com/xcconfig/

### Encrypt sensitive data

Install `gpg` and `git-crypt`:

```
brew install gpg
brea install git-crypt
```

Set up `git-crypt`:

```
git-crypt init
```

Create file `.gitattributes` and add those to it:

```
Keys/Keys.xcconfig filter=git-crypt diff=git-crypt
```

Commit files:

```
git add -A
git commit -m "Added .gitattributes for git-crypt"
```

Now if you push data to upstream, you'll find the `Keys.xcconfig` exists but encrypted.

### GPG keys

All collaborators should generate gpk key, including yourself.

```
gpg --generate-key
```

Configure your real name and email following the gpt instructions.

After the public and secret key created and signed, you will get something like:

```
pub   rsa2048 2019-01-07 [SC] [expires: 2021-01-06]
      D2B3EAAF9A8D5DB93CC30B26CCA243599CC80727B
uid           [ultimate] Your Name <your@email.com>
sub   rsa2048 2019-01-07 [E] [expires: 2021-01-06]
```

The second line is your key. You can also get it using the command `gpg --list-keys`.

### Add yourself to the git-crypt repo

Every collaborator needs to generate their own gpg key exactly the same way mentioned above. Let's add yourself first.

```
git-crypt add-gpg-user D2B3EAAF9A8D5DB93CC30B26CCA243599CC80727B
git push
```

After cloning the repo, you can unlock files by `unlock` command:

```
git-crypt unlock
```

### Other Collaborators

Other collaborators have to generate their own gpg keys on their own machine, export the public key and hand it over to you.

Ask them to export public key:

```
gpg --armor --export --output /Users/someuser/user_pubkey.gpg
```

After you get that file, import it:

```
gpg --import user_pubkey.gpg
```

Trust the key:

```
gpg ––edit–key D2B3EAAF9A8D5DB93CC30B26CCA243599CC80727B
```

Then you can add this collaborator using `git-crypt add-gpg-user xxxxx`

References:

* https://medium.com/@sumitkum/securing-your-secret-keys-with-git-crypt-b2fa6ffed1a6
* https://github.com/AGWA/git-crypt/wiki/Create-a-Repostiory
* http://irtfweb.ifa.hawaii.edu/~lockhart/gpg/

