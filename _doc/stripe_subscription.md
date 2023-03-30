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

3. 根据本次请求提供的参数，构建一个价格数据(程序中叫`CartItemStripe`)。

价格中包含了本次订阅使用到的所有价格的具体信息，这些数据可能来自内存的缓存、FTC数据库或者Stripe API（按顺序依次fallback）包含如下字段：

  * `Recurring` 是一个[StripePrice](./stripe_price.md)，这个是日常使用的价格，对应请求中的`priceId`字段

  * `Introductory` 也是[StripePrice](./stripe_price.md)，这是使用价格，对应请求中的`introductoryPriceId`

  * `Coupon` 是一个`StripeCoupon`

获取购物车数据后，如果存在`Introductory`，会做如下验证：

  * `Recurring`价格和`Introductory`价格必须数据同一个Stripe的product id，否则会认为无效；

  * `Introductory`的类别字段(`Kind`)必须为`one_time`，以防误把自动续订的价格关联成试用价格。

4. 此时获取到了用户账号、订阅产品的所有信息，构建一个购物车(`ShoppingCart`)，这个ShoppingCart是通用的，微信支付宝的一次性订阅也使用它，对于Stripe而言包含如下字段：

  *  `Account`用户账号基本数据，这里不包括订阅信息

  * `StripeItem` 即上一步中构建的价格数据

  * `PayMethod` 这里支付方式只有Stripe

5. 开始创建订阅，这里会锁表直到订阅创建成功或者失败。

6. 从`ftc_vip`中取出membership的数据，这里使用用户的compound id，即优先使用uuid，不存在则使用微信id；

7. 用membership数据更新上一步构建的购物车，这里会用当前的membership来推测出来本次订阅的意图，为防止各种订阅渠道之间的冲突，这里会做很多检测，具体检测过程见本文末尾。检测主要依赖当前会员状态的payment method数据，只有两种情况允许继续：

  * 过期且没有自动续订；

  * 微信支付宝订阅有效但想转为Stripe

其他情况都会被拒绝。

8. 创建订阅。注意，由于本程序分为live和sandbox两套环境运行（url不同），客户端发往不同地址的请求会分别使用Stripe的live/test API。

9. 订阅成功后构建一个新的数据结构，包括如下字段：

  * `Subs`: Stripe返回的订阅信息，精简后加上用户的uuid（注意Stripe订阅肯定是邮箱注册的用户），存储在`stripe_subscription`表;

  * `Member`: 新的会员信息，存储在`ftc_vip`表;

  * `Versioned`: 会员信息备份，包括了变更前和变更后的状态，这是不变的数据，以JSON格式存储;

  * `CarryOverInvoice`: 如果本次订阅前用户已经用微信支付宝购买了会员，并且还在有效期内，转为add-on，我们把剩余时间信息当作用户的一次零元购买，新建一条一次性购买invoice，存储在`ftc_invoice`表，留待以后重新使用.

至此锁表结束。

10. 存储本次购买过程相关参数备查，包括此前构建的购物车数据、购买前的会员状态、请求参数、Stripe返回的原始订阅信息，这里都是不变的数据，均以JSON列存储在`stripe_shopping_session`。

11. 如果使用了coupon，则生成一条`CouponRedeemed`数据存储在`stripe_coupon_redeemed`表，这里存储的是Stripe订阅的invoice id和coupon id的对应关系，以便记录用户是否在一个订阅周期内使用了任意一个coupon，以限制一次付款周期内重复使用coupon，例如，如果用户的自动续订是年度版，那么我们规定这一年内只能领取一次coupon，这个订阅一年生成一张invoice，那么只要某个invoice id和出现在这个表里，说明用户使用过了coupon，则这个付款周期内不用允许使用其他coupon。这里不能使用subscription id来记录，因为对于自动续订，不管多少年，这个id是始终不变的。

12. 把Stripe订阅原始数据下面携带的Invoice、Payment intent字段分别存储到`stripe_invoice`表和`stripe_payment_intent`表，仅供我方备份参考。

注意，10 ～ 12的存储过程均在goroutine中进行，9结束就马上向客户端返回数据了。

### Response

返回的数据是上述9中的`Subs`和`Membership`字段：

```json
{
    "subs": {
      "id": "stripe-subscription-id",
      "cancelAtUtc": null,
      "cancelAtPeriodEnd": false,
      "canceledUtc": null,
      "currentPeriodEnd": "2024-03-30T14:44:00Z",
      "currentPeriodStart": "2023-03-30T14:44:00Z",
      "customerId": "stripe-customer-id",
      "defaultPaymentMethod": "default-payment-method-id",
      "discount": null,
      "endedUtc": null,
      "ftcUserId": "uuid",
      "items": [
        {
          "id": "subscription-item-id",
          "price": {},
          "created": 1680159160,
          "quantity": 1,
          "subscriptionId": "stripe-subscription-id"
        }
      ],
      "latestInvoiceId": "latest-invoice-id",
      "liveMode": true,
      "paymentIntent": {
        "id": "payment-intent-id",
        "amount": 10,
        "cancellationReason": null,
        "clientSecret": null,
        "created": 1680159160,
        "currency": "gbp",
        "customerId": "stripe-customer-id",
        "invoiceId": "stripe-invoice-id",
        "liveMode": true,
        "paymentMethodId": "payment-method-id",
        "status": ""
      },
      "startDateUtc": null,
      "status": "active"
    },
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

`subs.items`是一个数组，一般有一个元素，如果有试用，则是introductory price加recurring price两个。

## Load a Subscription

```
GET /stripe/subs/{id}
```

`id` 是 [Subscription](https://stripe.com/docs/api/subscriptions/object) 的 `id` 字段。

客户端通常使用在本地保存的账号数据中的 `membership.stripeSubsId` 字段值。

### Response

See the subscription object at the start of this page.

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

See the subscription object.

## Refresh Subscription

```
POST /stripe/subs/{id}/refresh
```

### Response

See the subscription object.

## Cancel subscription

```
POST /stripe/subs/{id}/cancel
```

### Request

NULL

### Response

The subscription object.

## Reactivate subscription

```
POST /stripe/subs/{id}/reactivate
```

Reactivate a canceled subscription before its current period ends

### Response

The subscription object

## Load Subscription's Default Payment Method

```
GET /stripe/subs/{id}/default-payment-method?<refresh=true>
```

### Response

A payment method object:

```json
{
  "id": "payment-method-id",
  "customerId": "customer-id",
  "kind": "",
  "card": {
    "brand": "",
    "country": "",
    "expMonth": 3,
    "expYear": 2023,
    "fingerPrint": "",
    "funding": "credit",
    "last4": "1234",
    "network": {
      "available": ["amex", "dinners"],
      "preferred": "amex"
    }
  }
}
```

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

The Subscription object with `defaultPaymentMethod` field updated.

## Get Latest Invoice

```
GET /stripe/subs/{id}/latest-invoice
```

### Response

An `Invoice` object:

```json
{
    "id": "in_1KsLy7BzTK0hABgJbbOw0VgV",
    "accountCountry": "GB",
    "accountName": "THE FINANCIAL TIMES LIMITED",
    "amountDue": 3999,
    "amountPaid": 3999,
    "amountRemaining": 0,
    "attemptCount": 1,
    "attempted": true,
    "autoAdvance": false,
    "billingReason": "subscription_create",
    "chargeId": "ch_3KsLy7BzTK0hABgJ1s7fE4Sb",
    "collectionMethod": "charge_automatically",
    "currency": "gbp",
    "customerId": "cus_LZUxfDsrEGhd2h",
    "defaultPaymentMethod": "",
    "discounts": [],
    "hostedInvoiceUrl": "https://invoice.stripe.com/i/acct_1EpW3EBzTK0hABgJ/test_YWNjdF8xRXBXM0VCelRLMGhBQmdKLF9MWlV5a2g4dGVwWVMwVnVqTlVhdmpCeUdiZHRSZk5ELDQ5MjgxMTg00200ZEI67rqJ?s=ap",
    "invoicePdf": "https://pay.stripe.com/invoice/acct_1EpW3EBzTK0hABgJ/test_YWNjdF8xRXBXM0VCelRLMGhBQmdKLF9MWlV5a2g4dGVwWVMwVnVqTlVhdmpCeUdiZHRSZk5ELDQ5MjgxMTg00200ZEI67rqJ/pdf?s=ap",
    "liveMode": false,
    "nextPaymentAttempt": 0,
    "number": "1D6C8171-0001",
    "paid": true,
    "paymentIntentId": "pi_3KsLy7BzTK0hABgJ1963s1ot",
    "periodEndUtc": "2022-04-25T07:04:23Z",
    "periodStartUtc": "2022-04-25T07:04:23Z",
    "receiptNumber": "2686-5534",
    "status": "paid",
    "subscriptionId": "sub_1KsLy7BzTK0hABgJKv02yjaV",
    "total": 3999,
    "created": 1650870263
}
```

## Get coupon attached to latest invoice

Used to check if there's any coupon applied to a subscription's latest invoice. If it does, client should not show coupon to user.

```
GET /stripe/subs/{id}/latest-invoice/any-coupon
```

Response is a `CouponRedeemed` object. It always returns the same data structure with all fields set to zero values if not found.

```json
{
    "ftcId": "4eab2991-9669-41c2-b51e-75e1a0e76183",
    "invoiceId": "in_1KsLy7BzTK0hABgJbbOw0VgV",
    "liveMode": false,
    "subsId": "sub_1KsLy7BzTK0hABgJKv02yjaV",
    "couponId": "sj2iz99D",
    "createdUtc": "2022-07-25T09:13:04Z",
    "redeemedUtc": "2022-07-25T09:13:04Z"
}
```

Empty response:

```json
{
    "ftcId": "",
    "invoiceId": "",
    "liveMode": false,
    "subsId": "",
    "couponId": "",
    "createdUtc": null,
    "redeemedUtc": null
}
```

## Stripe订阅用户Intent的判断过程

* 对于过期用户(包括苹果过期且未开启自动续订)、Stripe无效的订阅，通常是新建订阅，这是最简单的情况；

* 如果payment method是微信、支付宝，则认为用户在把一次性购买转为自动续订，当前订阅的剩余时间会转为 add-on；

* 如果payment method是stripe，则有很多情况：

  * 如果会员的tier和本次订阅价格的tier、cycle均相同，而且存在coupon，则认为用户在兑换coupon，否则可能是重复订阅，禁止;

  * 如果会员的tier和本次订阅的tier相同、cycle不同，则认为用户在更改订阅周期;

  * 如果二者tier不同，请求价格的tier是premium，则认为用户在升级；如果请求的价格是standard，则认为用户在降级;

* 如果是payment method是apple，则禁止;

* 如果是B2B，则禁止.
