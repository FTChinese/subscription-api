# Introduction

## One-time-purchase

ftc_vip如下字段组合互斥：
payment_method == (alipay || wechat) 和 ftc_plan_id 同时存在 并且 auto_renewal为0
payment_method == apple  和 apple_subscription、auto_renewal 同时存在
payment_method == stripe 和 stripe_subscription_id、stripe_plan_id、sub_status、auto_renewal 同时存在

## Endpoints

* GET `/membership` Get a user's membership details
* PATCH `/membership` Update a user's membership. NOT implemented.
* PUT `/membership` Create a new membership. NOT implemented.
* GET `/membership/snapshots` Get a list of membership change history.
* POST `/membership/addons` Transfer addon to expiration time.

* GET `/orders` Get a list of orders of a user.
* GET `/orders/{id}` Show an order if a user.
* GET `/orders/{id}/payment-result` Get the payment result original send by Alipay/Wxpay as is.
* POST `/orders/{id}/verify-payment` Verify against Alipay/Wxpay whether an order is actually paid.

* GET `/invoices` Get a list of invoices of a user.
* PUT `/invoices` Create a new invoice. NOT implemented.
* GET `/invoices/{id}` Show the details of an invoice.

* GET `/stripe/prices?refresh=<true|false>` Get all Stripe active prices
* POST `/stripe/customer` Create a Stripe customer if not exists yet.
* GET `/stripe/customer/{id}` Get the details of a customer.
* POST `/stripe/customer/{id}/default-payment-method` Set the default payment method of a customer.
* POST `/stripe/customer/{id}/ephemeral-keys` Generate ephemeral key for client when it is trying to modify customer data.
* POST `/stripe/subs` Create a subscription.
* GET `/stripe/subs` List all subscriptions of a user.
* GET `/stripe/subs/{id}` Get a single subscription.
* POST `/stripe/subs/{id}` Update a subscription
* POST `/stripe/subs/{id}/refresh` Refresh a subscription.
* POST `/stripe/subs/{id}/cancel` Cancel a subscription.
* POST `/stripe/subs/{id/reactivate}` Reactivate a cancelled subscription.
