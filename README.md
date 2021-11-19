# Subscription API

## TOC

* [Getting Started](./_doc/getting_started.md)
* [Common Data Type Definition](./_doc/common_types.md)
* [Account Endpoints](./_doc/account_intro.md)
* [Ftc Purchase](./_doc/subscription_intro.md)
* [Stripe Subscription](./_doc/stripe_intro.md)
* [Apple Subscription](./_doc/apple_intro.md)
* [Database Design](./_doc/db_architecture.md)

API for subscription service

## Versioning

* v2: starting from January 2021.
* v3: starting from September 2021, due to discounts added.
* v4: Starting from November 2021, due to Stripe introductory offer.

When upgrading to a new version. do remember to change the following configurations manually:

1. In `pkg.config.port` file change the `Port` to a new one. Currently, all versions are using "820x" port. Replace the `x` with you current version.
2. In `pkg.config.webhook_url` file, change the production version of alipay/wechat pay server-to-server notification url.
3. Change stripe webhook url with the following steps:
   1. Go to Stripe dashboard.
   2. Find the "Developers" configuration page. 
   3. Select "Webhooks" from sidebar. 
   4. Click "Add endpoint". 
   5. Add new version's Stripe webhook url. 
   6. Copy the "Signing secret". 
   7. On your machine's configuration file `~/config/api.toml`. In the `api_keys` section, add a section like:

    ```toml
    [api_keys.stripe_webhook_v<your-current-version>]
    dev = "the test key. Simply copy it from previous versions since it won't be changed."
    prod = "the singing key you copied"
    ```
    8. Upload the configuration file to tk11 machine.
    9. Also run `make devconfig` command so that the modified configuration file is synced to your current directory so that Go's embedding of static assets works for development.
4. In makefile, change `app_name := subs-api-v<your version name>`
5. Configure URL. Map external URL to the new version's port.

| Version  | Port  | Binary name      | External URL |
| -------- | ----- | ---------------- | ------------ |
| 1        ｜ 8200 ｜ subscription-api | https://www.ftacademy.cn/api/v1 |
| 2        | 8202  | subs-api-v2      | https://www.ftacademy.cn/api/v2 |
| 3        | 8203  | subs-api-v3      | https://www.ftacademy.cn/api/v3 |
| 4        | 8204  | subs-api-v4      | https://www.ftacademy.cn/api/v4 |

## TODO

* Save alipay and wechat pay result to DB so that we could see the results without mining various tables. Every order has one row which reflects only the latest state. They should be updated every time a webhook payload received or order manually verified.  

## Build

This project produces 3 binaries: `subscription-api`, `subs_sandbox` and `iap-kafka-consumer`. By default, all make targets handles `subscripiton-api`. Set `APP` command line arguments to make to handle other binaries. The value could be one of `sandbox` or `consumer` for `subs_sandbox` and `iap-kafka-consumer` respectively.

```
make build APP=sandbox
make build APP=consumer
```

## Base URL

* Production: `http://www.ftacademy.cn/api/v<xx>`
* Sandbox: `http://www.ftacademy.cn/api/sandbox`

For Stripe and Apple IAP, the sandbox have the same meaning as their APIs: running in sandbox mode hits their sandbox endpoints; otherwise to production endpoints.

Sandbox means nothing to Wxpay and Alipay since they do not provide any means for testing.

To test these two payment providers, we use testing account. When you logged in with a testing account, the prices charged will be fixe to 0.01 and tell them to send webhook paywalod to the sandbox version.

Put it simply, if client detects the logged in account is of type test, send request to sandbox url; otherwise to production url. Only wxpay and alipay endpoints of this app will take further action for testing account. Stripe and Apple's will simply ignore it.

When you are using a testing account, you must also tell API your intention explicitly; otherwise the API won't check whether the request means testing and will treat it as a normal request. To do this, append query parameter `test=true` to wxpay and alipay endpoints. It won't have effects to other endpoints.

## Development

All development happens in the `sandbox` branch. `master` is used to merge and publish stable release from `sandbox`. The `sandbox` branch should always be test again live data to ensure subscription will never stop or break down.

## Issues

There's a problem with the wxpay sdk used here `github.com/objcoding/wxpay`. The `XmlToMap()` function this package provides does not take into account of indented or formatted XML. It requires that there's not space between each XML tag. If there are spaces and tabs, it cannot get the correct value.

## How to run

You need to put those files in the root directory of this project:

* `.env` Copy the `.env.example` file and rename it to `.env`
* `alipay_public_key.pem`
* `ftc_private_key.pem`

When deploying you need to put files exactly as follows:
```
go
 |----bin/
 |      |--- subscription-api
 |----.env
 |----alipay_public_key.pem
 |----ftc_private_key.pem
```

`cd` into directory `go` before run it since the program read those configuration files from the current working directory.

## Pitfalls

Wechat's APP ID, MCH ID must match, which means APP ID must be in your MCH ID's authorized list; otherwise you could never call wechat on you app.

32 char key is the one you entered, not generated by wechat.

## Renewal Policy

A subscribed user is only allowd to renew the next billing cycle. If the difference between current date and the expiration date is less than the billing cycle he is trying to subscribe, the user is allowed to renew subscription; otherwise, deny the request.

Example:

A user is subscribe to a yearly membership from 2018-01-01 to 2019-01-01. Today is 2018-07-01 and He is trying to extend one year more's subscription. The difference between toay (2018-07-01) to expiration date (2019-01-01) is less then a billing cycle (one year). He can renew subscription to 2020-01-01. After renewal, this user immediately tries to extend another billing cycle to 2021-01-01. This request will be denied since today (2018-07-01) to expiration date (2020-01-01) is greater than a yearly billing cycle.

The same priciple applies to monthly subscrition. A user could only buy two month consecutively at most.
