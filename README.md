# Subscription API

API for subscription service

## Base URL

* Production: `http://www.ftacademy.cn/api/v1`
* Sandbox: `http://www.ftacademy.cn/api/sandbox`

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
