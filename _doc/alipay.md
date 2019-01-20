# Alipay

## Create Alipay Order

    POST /alipay/app-order/{tier}/{cycle}

### Input

`tier` must be one of `standard` or `premium`.

`cycle` must be one of `year` or `month`.

Request header must contain `X-User-Id: <uuid>` if user logged in with FTC account or `X-Union-Id: <wechat union id>` if used logged in with Wechat OAuth. If an FTC account is already bound to a Wechat account, and user logged in via Wechat OAuth, you **MUST** always use `X-User-Id: <uuid>` and should never use `X-Union-Id`.

Request header should also contain:
```
X-Client-Type: <web|ios|android>
X-Client-Version: <major.minor.patch>
```

If `X-Client-Type` is `web`, the request header must also contain
```
X-User-Agent: <forwarded user agent for web> 
X-User-Ip: <forwareded user ip>
```

### Response

* `401 Unauthorized` if request header does not contain `X-User-Id`.

* `400 Bad Request`

If `tier` and `cycle` is not one of the values as specified above;

If subscription plan if not found;

If sign request parameters failed;

* `403 Forbidden`

If this user is already a member and current date is not within the allowed renewal period.

```json
{
    "message": "Already a subscribed user and not within allowed renewal period.",
}
```

* `404 Not Found` if current does not exist.

* `200 OK`

```json
{
    "ftcOrderId": "string",
    "netPrice": 258,
    "listPrice": 258,
    "param": "string" // Pass this string to alipay sdk.
}
```