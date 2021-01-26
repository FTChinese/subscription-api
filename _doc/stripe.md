# Stripe订阅

Stripe订阅仅限使用邮箱登录的用户。用户ID通过 HTTP header `X-User-Id` 字段设置。
除 `/stripe/prices` 之外的请求均需提供该值。

## Customer

在使用Stripe订阅之前，必须先在Stripe[创建 Customer](https://stripe.com/docs/api/customers/create).

## Stripe订阅状态

参见 https://stripe.com/docs/api/subscriptions

* `incomplete`
* `incomplete_expired`
* `trialing`
* `active`
* `past_due`
* `canceled`
* `unpaid`

状态转化如下：

```
                 [not paid in 23 hours]
     incomplete -----------------------> incomplete_expired 
    |         \ 
  [failed]     \ [paid]
    |           \               /--> past_due
(initial pay)    | ---> active | ---> canceled
                /               \--> unpaid
               /
  trialing    /
```

## Cancel

新订阅创建时默认自动扣款，取消订阅即取消自动扣款。有三种方式取消：

[Subscription](https://stripe.com/docs/api/subscriptions/object)

* `cancel_at`
* `cancel_at_period_end`
* `canceled_at`
* `current_period_end`
* `current_period_start`
* `status`

1. 删除订阅

Stripe API URL: `DELETE /v1/subscription/:id`. 

订阅的`canceled_at`字段会设为取消的日期，`status`=`canceled`。该订阅即刻取消，Stripe不再接受对改订阅的任何修订。

```json
{
  "cancel_at": null,
  "cancel_at_period_end": false,
  "canceled_at": 1611625523,
  "current_period_end": 1638943057,
  "current_period_start": 1607407057,
  "status": "canceled"
}
```

参考 [Cancel a subscription](https://stripe.com/docs/api/subscriptions/cancel).

2. 在本次订阅结束时不再自动扣款 

Stripe API URL: `POST /v1/subscription/:id`.

Request body:

```
cancel_at_period_end=true
```

Example Result:

```json
{
  "cancel_at": 1643161258,
  "cancel_at_period_end": true,
  "canceled_at": 1611625735,
  "current_period_end": 1643161258,
  "current_period_start": 1611625258,
  "status": "active"
}
```

请求成功后，`cancel_at_period_end`变为`true`, `cancel_at` (2022-01-26 01:40:58) 设为 `current_period_end`相同的值, `status` 仍是 `active`. `canceled_at` (2021-01-26 01:48:55) 则表示执行改操作的时间.

See [Update a subscription](https://stripe.com/docs/api/subscriptions/update).

3. 在指定时间取消

Stripe API URL `POST /v1/subscription/:id`.

Request Body:

```
cancel_at=1643161258
```
可以是过去或未来的时间。

我们只使用第二种方式。

### 找出过期时间

综合 `status`、`canceld_at`、`cancel_at_period_end` 三个字段决定会员过期时间和自动续订状态：

```
var autoRenew: bool
var expirationDate: Date

if status == "canceled" {
    autoRenew = false

    if !cancel_at && !cancel_at_period_end {
        expirationDate = canceled_at
        return        
    }
    
    if cancel_at_period_end {
        expirationDate = current_period_end
        return
    }

    expirationDate = cancel_at
}
```

## 处理Stripe的API接口列表

* `/stripe/prices`
* `/stripe/customers`
* `/stripe/customers/{id}`
* `/stripe/customers/{id}/default-payment-method`
* `/stripe/customers/{id}/ephemeral-keys`
* `/stripe/subs`
* `/stripe/subs/{id}`
* `/stripe/subs/{id}/refresh`
* `/stripe/subs/{id}/upgrade`
* `/stripe/subs/{id}/cancel`
* `/stripe/subs/{id}/reactivate`

注意：API返回的数据格式并不是Stripe的原始数据格式。

API在调用Stripe SDK时，遇到Stripe API返回的任何错误，都原样转发给客户端，不作改动，见 [Errors](https://stripe.com/docs/api/errors)。

## 获取Stripe价格列表

```
GET /stripe/prices?refresh=true|false
```

改请求带有一个可选参数 `refresh`，值为字符串 `true` 或 `false`。用户强制从Stripe API获取数据，否则返回可能是缓存数据。

### Workflow

1. 解析URL的表单参数。如果无法解析，返回 `400 Bad Request`。仅接受字符串`"true"` 为 `true`，其他所有值被当作 `false`。
2. 如果 `refresh` 不是 `true`，且缓存中存在数据， 则返回缓存数据
3. 否则，从 Stripe [List all prices](https://stripe.com/docs/api/prices/list) 获取数据
4. 缓存数据
5. 返回数据

### Example response:

```json
[
    {
        "id": "plan_FOdgPTznDwHU4i",
        "tier": "standard",
        "cycle": "month",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Standard Monthly Plan",
        "productId": "prod_FOde1wE4ZTRMcD",
        "unitAmount": 390,
        "created": 1562567567
    },
    {
        "id": "plan_FOdfeaqzczp6Ag",
        "tier": "standard",
        "cycle": "year",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Standard Yearly Plan",
        "productId": "prod_FOde1wE4ZTRMcD",
        "unitAmount": 3000,
        "created": 1562567504
    },
    {
        "id": "plan_FOde0uAr0V4WmT",
        "tier": "premium",
        "cycle": "year",
        "active": true,
        "currency": "gbp",
        "liveMode": false,
        "nickname": "Premium Yearly Plan",
        "productId": "prod_FOdd1iNT29BIGq",
        "unitAmount": 23800,
        "created": 1562567431
    }
]
```


## 创建Stripe用户

```
POST /stripe/customers
```

### Workflow

1. 从HTTP Header中获取UUID。

2. 锁表

3. 从数据库取出用户账号。未找到返回 `404 Not Found`

4. 如果数据中 `stripe_customer_id` 字段存在，则认为该账号已经注册了Stripe用户，使用该值从Stripe获取到用户数据，返回，结束。

5. 如果尚未注册Stripe用户，调用Stripe SDK创建用户，提交的数据中仅包含用户的邮箱。

6. 保存 Customer ID 到 `stripe_customer_id` 字段。

遇到的任何数据库错误返回 `500 Internval Server Error`

### Example Response

```json
{
    "id": "cus_IXp31Fk2jYJmU3",
    "ftcId": "c07f79dc-664b-44ca-87ea-42958e7991b0",
    "defaultSource": null,
    "defaultPaymentMethod": "pm_1Hzzx3BzTK0hABgJGy155ZR1",
    "email": "stripe.test@ftchinese.com",
    "liveMode": false,
    "createdUtc": "2020-12-10T07:17:54Z"
}
```


## 获取Stripe Customer的详情

```
GET /stripe/customers/{id}
```

ID是Stripe customer的id。改请求仍需提供FTC的UUID，用于验证二者是否一致。

### Workflow

1. 从HTTP header获取UUID。

2. 获取URL中的 `id` 值。

3. 使用UUID获取用户账号数据

4. 检查 `stripe_customer_id` 是否为空。如果是，则该 Stripe customer 不存在，返回404。

5. 检查 `stripe_customer_id` 是否与 `id` 相同，不相同返回404。此举旨在防止客户端发送的数据错误导致FTC用户A获取到了FTC用户B的Stripe数据。

6. 从Stripe API获取customer数据。错误和返回数据同上一节。

## 更新默认支付方式

```
POST /stripe/customers/{id}/default-payment-method
```

### Request Body

```json
{
  "defaultPaymentMethod": "id of a payment method"
}
```

### Workflow

1. 从header获取uuid，url获取customer id，request body获取请求参数。

2. 如果request body无法解析，返回 `400 Bad Request`；如果数据不合法，返回 `422 Unprocessable`:

```json
{
  "message": "Missing required field",
  "error": {
    "field": "defaultPaymentMethod",
    "code": "missing_field"
  }
}
```

3. 使用uuid取出用户账号。未找到、账号不存在stripe customer id，或不等于url中的id，返回404。

4. 发送请求到Stripe。

5. 返回数据同上一节。

## 获取Ephemeral Keys

```
POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
```

URL参数 `api_version` 为必填项，从客户端SDK中获取。

Stripe API的数据原样返回，客户端SDK直接使用。


## 新建订阅

```
POST /stripe/subs
```

注意：Stripe API严格按照restful方式设计，它没有状态，不会检查用户是不是有多个订阅，只要请求发送到 `https://api.stripe.com/v1/subscriptions` 就会创建一个订阅，因此，我们的API和客户端都需要检测用户是不是会产生多次订阅的情况。

### Request Body

```json
{
  "tier": "standard | premium, required",
  "cycle": "year | month, required",
  "coupon": "stripe coupon id, optional",
  "defaultPaymentMethod": "payment method id, required",
  "idempotency": "a unique string client generated to prevent duplicate request"
}
```

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
    "subs": {
        "id": "sub_IpPj35ZNbBLLsP",
        "tier": "standard",
        "cycle": "year",
        "cancelAtUtc": null,
        "cancelAtPeriodEnd": false,
        "canceledUtc": null,
        "currentPeriodEnd": "2022-01-26T06:19:51Z",
        "currentPeriodStart": "2021-01-26T06:19:51Z",
        "customerId": "cus_IXp31Fk2jYJmU3",
        "defaultPaymentMethod": null,
        "subsItemId": "si_IpPj7tQVLU1OZv",
        "priceId": "plan_FOdfeaqzczp6Ag",
        "latestInvoiceId": "in_1IDku3BzTK0hABgJeoPu2E3v",
        "liveMode": false,
        "startDateUtc": "2021-01-26T06:19:51Z",
        "endedUtc": null,
        "createdUtc": "2021-01-26T06:19:51Z",
        "updatedUtc": "2021-01-26T06:19:53Z",
        "status": "active",
        "ftcUserId": "c07f79dc-664b-44ca-87ea-42958e7991b0"
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

如果`payment.requireAction == true`，可能用户的卡需要验证，按Stripe要求向客户端SDK提供`paymentIntentClientSecret`的值。

`subs.id` 和 `membership.stripeSubsId` 都可以用来获取 `subs` 字段，见下。

## 列出一个用户的所有订阅

```
GET /stripe/subs
```

未实现

## 获取用户的某个订阅详情

```
GET /stripe/subs/{id}
```

`id` 是 [Subscription](https://stripe.com/docs/api/subscriptions/object) 的 `id` 字段。

客户端通常使用在本地保存的账号数据中的 `membership.stripeSubsId` 字段值。

### Response

```json
{
    "id": "sub_IX3JAkik1JKDzW",
    "tier": "premium",
    "cycle": "year",
    "cancelAtUtc": null,
    "cancelAtPeriodEnd": false,
    "canceledUtc": null,
    "currentPeriodEnd": "2021-12-08T05:57:37Z",
    "currentPeriodStart": "2020-12-08T05:57:37Z",
    "customerId": "cus_FRgIy7R6sn5nI7",
    "defaultPaymentMethod": null,
    "subsItemId": "si_IX3JJ7rrB8wmwY",
    "priceId": "plan_FOde0uAr0V4WmT",
    "latestInvoiceId": "in_1HvzCfBzTK0hABgJklW9Azi5",
    "liveMode": false,
    "startDateUtc": "2020-12-08T05:57:37Z",
    "endedUtc": null,
    "createdUtc": "2020-12-08T05:57:37Z",
    "updatedUtc": "2020-12-10T03:51:25Z",
    "status": "active",
    "ftcUserId": "8637a4b2-098d-4f42-9fb4-a5edeae2309d"
}
```


## 标准版升级高端版

```
POST /stripe/subs/{id}/upgrade
```

和新建订阅基本类似，除了请求参数中 `tier` 字段必须等于 `premium`.

## 关闭自动续订

```
POST /stripe/subs/{id}/cancel
```

无请求参数

### Workflow

1. 从header获取uuid，url获取订阅id。

2. 使用uuid找到用户的membership，验证该会员付费方式是否来自Stripe，以及 `stripeSubsId` 是否等于url中传入的id。

3. 如果 `membership.autoRenew == false`，则用户已经关闭了自动续订，结束。返回的数据同新建订阅，但只有 `membership` 字段有值，`subs` 和 `payment`尽管也存在，但是它们下面的字段都是空值。由于我们在Golang尽量避开使用指针，因此复合数据类型很难产生 `null` 值，其下的原始数据类型都使用了zero value。关于Golang的zero value，见 https://tour.golang.org/basics/12。

4. 发送请求到Stripe，在本次订阅结束时取消本订阅。详见Stripe API [Update a subscription](https://stripe.com/docs/api/subscriptions/update) 中的 `cancel_at_period_end` 字段介绍。

5. 结束。返回的数据结构同新建会员。


## 重新打开自动续订

```
POST /stripe/subs/{id}/reactivate
```

与关闭自动续订完全相同，只不过发到Stripe API的请求是 `cancel_at_period_end=true`.
