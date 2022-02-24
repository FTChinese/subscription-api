# Stripe订阅

Stripe订阅仅限使用邮箱登录的用户。用户ID通过 HTTP header `X-User-Id` 字段设置。
除 `/stripe/prices` 之外的请求均需提供该值。

## Index

* [Customers](./stripe_customers.md)
* [Setup Intent](./stripe_setup_intent.md)
* [Payment Method](./stripe_payment_method.md)
* [Subscription](./stripe_subscription.md)

## API List

* Get `/stripe/prices` List stripe prices
* Get `/stripe/prices/{id}` Get a specific price
* POST `/stripe/customers` Create a new customer
* GET `/stripe/customers/{id}` Get a customer details
* GET `/stripe/customers/{id}/default-payment-method` Get the default payment method set on customer
* POST `/stripe/customers/{id}/default-payment-method` Set/Change customer's default payment method
* GET `/stripe/customers/{id}/payment-methods` List a payment methods of a customer.
* POST `/stripe/customers/{id}/ephemeral-keys` Create an ephemeral key when native app wants to modify customer data.
* POST `/stripe/setup-intents` Create a setup intent
* GET `/stripe/setup-intents/{id}?<rfresh=true>` Get a setup intent
* GET `/stripe/setup-intents/{id}/payment-method` The payment method created by this setup intent.
* GET `/stripe/payment-methods/{id}?<refresh=true>` Get a payment method
* POST `/stripe/subs` Create a new subscription
* GET `/stripe/subs/{id}` Get a subscription
* POST `/stripe/subs/{id}` Update a subscription
* POST `/stripe/subs/{id}/refresh` Refresh a subscription
* POST `/stripe/subs/{id}/cancel` Cancel a subscription
* POST `/stripe/subs/{id}/reactivate` Undo a scheduled cancellation of subscription.
* GET `/stripe/subs/{id}/default-payment-method` Get the default payment method of a subscription.

注意：

* API返回的数据格式并不是Stripe的原始数据格式。
* Customer和Subscription的default payment method不一定一致。

API在调用Stripe SDK时，遇到Stripe API返回的任何错误，都原样转发给客户端，不作改动，见 [Errors](https://stripe.com/docs/api/errors)。

## Usage Guide

You always start using Stripe by creating [Products](https://stripe.com/docs/api/products). 

## Design of introductory offer

Take advantage of [Combining trials with add_invoice_items](https://stripe.com/docs/billing/subscriptions/trials#combine-trial-add-invoice-items).

The general idea is that when creating a subscription, you should set a trial period which is free by Stripe's definition. Then add an extra invoice for a specific one-time price.

Steps to set up an introductory offer:

* Open the product page you want to attach an introductory offer. 
* Click "Add another price".
* Enter a price you want to charge for this introductory offer. Select "One time".
* In the "Price description" box, enter something that could help you recall its purpose.
* After the price created, click it and "Edit metadata". Enter the following key-value pairs:
    - `tier`: The tier of product this price belongs to. For standard edition it's `standard`, or `premium` for premium edition.
    - `period_days`: The number of days for this introductory offer. For example `7` for a week's trial, `30` or `31` or a month of trial.
    - `introductory`: `true`.
    - `start_utc`: A UTC time string when this price is available. For example `2021-11-16T16:00:00Z`
    - `end_utc`: A UTC time string when this price should be ended. For example `2021-12-16T16:00:00Z`

Once created, the introductory offer will be kept forever unless you deleted it. However, it won't come into effect forever. It will only be used when the paywall data fetch from this API contains a valid introductory offer under a price. So as long as you set any introductory offer under a price in Superyard, alipay, wechat pay, and Stripe will all be notified to offer user an introductory trial period.

Do remember to set the "Edit metadata" as stated above. I need this information to distinguish a introductory price from regular prices. You should also keep in mind not creating multiple introductory price under the same product; otherwise there's currently no way to determine which one to use.

### Upgrade from standard to premium during trialing period

Stripe allows a subscription switching another pricing, which applies to trialing period. If a user is subscribed to standard edition with a trialing period, it could switch to premium directly. 

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

## Gotchas and Pitfalls

### Webhook

You could create as many webhook endpoints as you like. However, every webhook endpoint has its own signing key. Do remember to update it when you change webhook.
