## Endpoints

### Payment

Required headers: `X-User-Id` or `X-Union-Id`, with `X-User-Id` preferred if exists.

* POST `/wxpay/app` Payment in native app.
* POST `/wxpay/desktop` Payment in desktop web browser
* POST `/wxpay/mobile` Payment in mobile web browser
* POST `/wxpay/jsapi` Payment inside Wechat-embedded browser.
* POST `/alipay/desktop` Create a payment intent for desktop browser
* POST `/alipay/mobile` Create a payment intent for mobile browser
* POST `/alipay/app` Create a payment intent for native app.

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
