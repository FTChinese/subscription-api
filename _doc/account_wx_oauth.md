# Wechat Login and Binding

## Wechat Login for Native App

    POST /wx/oauth/login

Client get OAuth 2.0's `code` from Wechat API and sent the `code` to this endpoint.

Request header should contain those fields:
```
X-Client-Type: "web | ios | android"
X-Client-Version: "1.2.0"
```

For `X-Client-Type` == `web`, those field should also be provided:

```
X-User-Ip: "127.0.0.1"
X-User-Agent: "Mozilla"
```

Client request header **MUST** also include field `X-App-Id`, which is the value of Wechat app id. This is used to identify which app id and secret to use to forward request to Wechat API.

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
    "sessionId": "string",
    "unionId": "string",
    "createdAt": "2019-01-17T107:36:20Z"
}
```

`sessionId` is the hash value of MD5. It is generated from a string of pattern `<Access Token>:<Refresh Token>:<Open ID>`. Since Wechat gives each login attempt different access token, refresh token, and the access  token and refresh token are forbidden from being stored in client side, we create this hash value to find this session in the future.

Client should, after acquiring the session data, use `unionId` to send request to next-api's `/wx/account` to get a user's full account data.

Since refresh token has an expiration period of 30 days, client should also caculate if the session is expired and asks user to re-authenticate upon the end of the 30 days.

## Refresh Wechat User Info

    PUT `/wx/oauth/refresh`

Request header must include `X-App-Id` which is the wechat app you use to perform Wechat login, and `X-Session-Id`.

### Response

* `400 Bad Request`
* `204 No Content` if user info is refreshed and client should retrieve user account data again.

## Forward `code` for Web Client

    GET `/wx/oauth/callback?code=<code>&state=<random string>`

This endpoint is used to received Wechat response, not used by us directly. Upon receiving the OAuth `code`, it will redirect the back to web application. This is a way to circumvent Wechat retriction on authorized callback URL.

