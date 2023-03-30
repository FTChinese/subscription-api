## Subscription

The Subscription object:

```json
{
  "id": "xxx",
  "tier": "standard",
  "cycle": "year",
  "cancelAtUtc": null,
  "cancelAtPeriodEnd": false,
  "canceledUtc": null,
  "currentPeriodEnd": "2023-01-22T03:34:02Z",
  "currentPeriodStart": "2022-01-22T03:34:02Z",
  "customerId": "xxx",
  "defaultPaymentMethod": "xxx",
  "endedUtc": null,
  "ftcUserId": "xxx",
  "items": [
    {
      "id": "xxx",
      "price": {
        "id": "xxx",
        "active": true,
        "currency": "gbp",
        "isIntroductory": false,
        "kind": "recurring",
        "liveMode": false,
        "nickname": "Regular Yearly Charge",
        "productId": "xxx",
        "periodCount": {
          "years": 1,
          "months": 0,
          "days": 0
        },
        "tier": "standard",
        "unitAmount": 3999,
        "startUtc": null,
        "endUtc": null,
        "created": 1613617385,
        "product": "prod_FOde1wE4ZTRMcD",
        "recurring": {
          "interval": "year",
          "intervalCount": 1,
          "usageType": "licensed"
        },
        "type": "recurring"
      },
      "created": 1640144042,
      "quantity": 1,
      "subscriptionId": "xxx"
    }
  ],
  "latestInvoiceId": "xxx",
  "liveMode": false,
  "paymentIntent": {
    "id": "",
    "amount": 0,
    "cancellationReason": "",
    "clientSecret": null,
    "currency": "",
    "customerId": "",
    "invoiceId": "",
    "liveMode": false,
    "next_action": {},
    "paymentMethodId": "",
    "status": ""
  },
  "startDateUtc": "2021-12-22T03:34:02Z",
  "status": "active"
}
```

## 新建订阅

```
POST /stripe/subs
```

注意：Stripe API严格按照restful方式设计，它没有状态，不会检查用户是不是有多个订阅，只要请求发送到 `https://api.stripe.com/v1/subscriptions` 就会创建一个订阅，因此，我们的API和客户端都需要检测用户是不是会产生多次订阅的情况。

### Request Body

```json
{
  "priceId": "stripe price id",
  "introductoryPriceId": "stripe price id",
  "coupon": "stripe coupon id, optional",
  "defaultPaymentMethod": "payment method id, required",
  "idempotency": "a unique string client generated to prevent duplicate request"
}
```

* `priceId: string` Required
* `introductoryPriceId?: string` Optional
* `coupon?: string` Optional, Stripe coupon id.
* `defaultPaymentMethod: string` Optional but recommended.
* `idempotency?: string` Optional. Only required for Android SDK.

`introductoryPriceId`和`coupon`是互斥的，一个人享用intro price的时候，是不应该有coupon的。

当coupon存在的时候，既可以用于新订阅，也可以用在现有订阅。但是请注意，coupon一定是从客户端发起的。对于新用户，在订阅的时候就默认带上，但是对于已有用户则不同，由于已有用户默认是Stripe从银行直接扣款，因此，coupon需要用户在订阅界面点击额外的领取按钮，更新订阅，所以coupon的使用对于新老用户而言是不同的操作。如果在coupon存在期间，现有订阅用户没有去订阅界面领取，那么是不可能使用本次coupon的。

### Workflow

1. 从header获取uuid，解析request body并验证。

2. 取出用户账号，检查Stripe customer id是否存在。

3. 验证账号是否与当前运行环境一直。测试账号仅允许用于sandbox环境。测试账号指 `****.test@ftchinese.com` 类型的邮箱。

4. 使用请求中的 `tier` 和 `cycle` 找到 Stripe pricing 的 id。

5. 从 `ftc_vip` 表中找出用户会员信息，不存在的会员为合法值，等同于过期会员。

6. 检测当前会员是否可以订阅请求的产品：

```
function allowNewSubs(membership, tier, cycle): boolean {
    if membership.isEmpty {
        return true
    }
    
    if membership.isExpired {
        return true
    }
    
    if membership.paymentMethod != "stripe" {
        return false
    }
    
    if ["active", "incomplete", "trialing"] not include membership.status {
        return true
    }
    
    return false
}
```

7. 请求Stripe API创建订阅，根据Stripe订阅生成会员信息。

### Response

```json
{
    "payment": {
        "requiresAction": false,
        "paymentIntentClientSecret": "pi_1IDku3BzTK0hABgJtDduTOfL_secret_98WBq0pbpwMMPmYI9eTz24Vts"
    },
    "subs": {},
    "membership": {
        "ftcId": "c07f79dc-664b-44ca-87ea-42958e7991b0",
        "unionId": null,
        "tier": "standard",
        "cycle": "year",
        "expireDate": "2022-01-26",
        "payMethod": "stripe",
        "stripeSubsId": "sub_IpPj35ZNbBLLsP",
        "autoRenew": true,
        "status": "active",
        "appleSubsId": null,
        "b2bLicenceId": null,
        "standardAddOn": 0,
        "premiumAddOn": 0
    }
}
```

* `subs` is a Subscription object.

如果`payment.requireAction == true`，可能用户的卡需要验证，按Stripe要求向客户端SDK提供`paymentIntentClientSecret`的值。

`subs.id` 和 `membership.stripeSubsId` 都可以用来获取 `subs` 字段，见下。

## Load a Subscription

```
GET /stripe/subs/{id}
```

`id` 是 [Subscription](https://stripe.com/docs/api/subscriptions/object) 的 `id` 字段。

客户端通常使用在本地保存的账号数据中的 `membership.stripeSubsId` 字段值。

### Response

`Subscription`.

## Update a Subscription

```
POST /stripe/subs/{id}
```

Change a subscription to other price.

### Request

```json
{
  "priceId": "required",
  "coupon": "optional",
  "defaultPaymentMethod": "optional",
  "idempotency": "optional string"
}
```

* `priceId` is the price user wants to change to
* `defaultPaymentMethod` an existing subscription is not required to provide a payment method.

### Response

Save as creating new subscription.

## Refresh Subscription

```
POST /stripe/subs/{id}/refresh
```

### Response

Save as creating new subscription

## Cancel subscription

```
POST /stripe/subs/{id}/cancel
```

### Request

NULL

### Response

Save as creating new subscription

## Reactivate subscription

```
POST /stripe/subs/{id}/reactivate
```

Reactivate a canceled subscription before its current period ends

### Response

Save as creating new subscription.

## Load Subscription's Default Payment Method

```
GET /stripe/subs/{id}/default-payment-method?<refresh=true>
```

### Response

`Subscription`

## Update Subscription's Default Payment Method

```
POST /stripe/subs/{id}/default-payment-method
```

### Request

```json
{
  "defaultPaymentMethod": "required string"
}
```

### Response

The `Subscription` object

