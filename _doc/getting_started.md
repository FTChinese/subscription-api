# Getting Started

## Base URL

```
https://www.ftacademy.cn/api/v2
```

Using the v2.x.x of this repository.

`https://www.ftacdamy.cn/api/v1` uses v1.x.x of this repository and is kept only for backward compatibility in case some users didn't upgrade Android app timely. The service might be stopped in the future.

## Authorization Header

To impose access restriction, all requests are required to set header `Authorization: Bearer <token>`; otherwise the request is denied.

The only exceptions are endpoints under `/webhook` which are used by other parties to send notification.

## Internal Status

* `/__version` See current program's build info

