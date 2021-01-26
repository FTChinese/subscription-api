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

* GET `/stripe/prices` 获取Stripe价格列表
* POST `/stripe/customers`
* GET `/stripe/customers/{id}` 获取Stripe Customer的详情。ID时Stripe customer的id
* POST `/stripe/customers/{id}/default-payment-method`
* POST `/stripe/customers/{id}/ephemeral-keys`
* POST `/stripe/subs` 创建订阅
* GET `/stripe/subs` 获取某用户的所有订阅
* GET `/stripe/subs/{id}` 获取某个用户的某个订阅
* POST `/stripe/subs/{id}/refresh` 刷新某个用户的某个订阅
* POST `/stripe/subs/{id}/upgrade` 升级某个用户的标准版订阅到高级版
* POST `/stripe/subs/{id}/cancel` 关闭某用户的订阅的自动续订
* POST `/stripe/subs/{id}/reactivate` 对于用户已经关闭自动续订但本次订阅未到期的订阅，重新打开自动续订



