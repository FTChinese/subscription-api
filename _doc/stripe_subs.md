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
