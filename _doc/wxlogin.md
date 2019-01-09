# Wechat Login and Binding

## Wechat Login

    POST /oauth/wx-access

Client get OAuth 2.0's `code` from Wechat API and sent the `code` to this endpoint.

Request header should contain those fields:
```
X-Client-Type: "web | ios | android"
X-Client-Version: "1.2.0"
```

For `X-Client-Type` == `web`, those field should also be provied:

```
X-User-Ip: "127.0.0.1"
X-User-Agent: "Mozilla"
```

TODO: include wechat app id in request header so that API knows which app id to use to send request to wechat. If app id is mismatched between API and client, the `code` will be regarded as invalid by wechat.

### Input

```json
{
    "code": "001hPJNE0xvMjl25EaLE0k4PNE0hPJNl"
}
```

### Response

* `400 Bad Request` if input JSON cannot be parsed.

```json
{
    "message": "Problems parsing JSON"
}
```

Or if error occurred while sending request to Wechat API.

* `422 Unprocessable Entity` if `code` is empty.

```json
{
    "message": "Validation failed",
    "error": {
        "field": "code",
        "code": "missing_field"
    }
}
```

* `200 OK` if Wechat userinfo is retrieved from Wecha API:
```json
{
    "unionID": "",
    "openID": "",
    "nickName": "",
    "avatarUrl"
}
```

This is the bare-bone data to identify a wechat user. After this step, you should generally request for this wechat user's full account data as specified by FTC. See the next step.

