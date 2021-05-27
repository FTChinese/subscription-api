# Introduction

## Endpoints

### One-time-purchase

Required headers: `X-User-Id` or `X-Union-Id`, with `X-User-Id` preferred if exists.

* POST `/wxpay/app` Payment in native app.
* POST `/wxpay/desktop` Payment in desktop web browser
* POST `/wxpay/mobile` Payment in mobile web browser
* POST `/wxpay/jsapi` Payment inside Wechat-embedded browser.
* POST `/alipay/desktop` Create a payment intent for desktop browser
* POST `/alipay/mobile` Create a payment intent for mobile browser
* POST `/alipay/app` Create a payment intent for native app.

### Membership

Required headers: `X-User-Id` or `X-Union-Id`, or both for linked account.

* GET `/membership` Get a user's membership details
* PATCH `/membership` Update a user's membership. NOT implemented.
* PUT `/membership` Create a new membership. NOT implemented.
* GET `/membership/snapshots` Get a list of membership change history.
* POST `/membership/addons` Transfer addon to expiration time.

### Orders

Required headers: `X-User-Id` or `X-Union-Id`, or both for linked account.

* GET `/orders` Get a list of orders of a user.
* GET `/orders/{id}` Show an order if a user.
* GET `/orders/{id}/payment-result` Get the payment result original send by Alipay/Wxpay as is.
* POST `/orders/{id}/verify-payment` Verify against Alipay/Wxpay whether an order is actually paid.

### Invoices

Required headers: `X-User-Id` or `X-Union-Id`, or both for linked account.

* GET `/invoices` Get a list of invoices of a user.
* PUT `/invoices` Create a new invoice. NOT implemented.
* GET `/invoices/{id}` Show the details of an invoice.

### Stripe

Headers required: `X-User-Id` should provide ftc's uuid except the `/stripe/prices`.

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

### Apple In-App-Purchase

No `X-User-Id` header is required unless explicitly specified.

The `{id}` placeholder here refers to Apple's original transaction id.

* POST `/apple/verify-receipt` Verify a receipt and returns the latest version.
* POST `/apple/link` Link email account to Apple subscription
* POST `/apple/unlink` Unlink ftc account from Apple subscription
* POST `/apple/subs` Verify a receipt and returns a condensed version of Apple's response.
* GET `/apple/subs` Get a list of a user's subscription. `X-User-Id` is required to identify the user.
* GET `/apple/subs/{id}` Load a single subscription
* PATCH `/apple/subs/{id}` Refresh an existing subscription.
* GET `/apple/recept/{id}` Load a receipt and its associated subscription.

### Paywall

* GET `/paywall` Loads the paywall data
* GET `/paywall/plans` Loads all active plans.
* GET `/paywall/__refresh` Bust cache since in the above two steps data are cached in-memory once data retrieved from DB, forever.

### Webhook

Handle payment provider's server-to-server notification.

* POST `/webhook/wxpay` Wechat pay notification
* POST `/webhook/alipay` Alipay notification
* POST `/webhook/stripe` Stripe notification
* POST `/webhook/apple` Apple notification

### Server status

* GET `/__version` Current version of this API.

## One-time-purchase

TODO

## Subscription

TODO
