# 邮箱或手机账号

## Load account

```
GET /account
```

## Delete account

```
DELETE /account
```

Delete account created with email or mobile. An account is allowed to be removed only when:

* This account must be created at ftchinese. For example, wechat account is not created at ftchinese.
* The account must not have a valid membership at the moment the deletion is performed.
* If the target account does have a valid subscription, user should email request deletion manually.

### Request Header

```
X-User-Id: string
```

### Request body

```json
{
  "email": "string",
  "password": "string"
}
```

We need to verify user's credentials before proceeding to deletion.

### Errors

* If request body fields have error

422 Unprocessable:

```json
{
  "message": "Invalid email",
  "error": {
    "field": "email | password",
    "code": "invalid | missing_field"
  }
}
```

* If password mismatched: 403 Forbidden
* If we failed to find account by this id: 404 Not Found
* If the email in request body does not match the account's email field we retrieved by the header id: 422 Unprocessable

```json
{
  "message": "",
  "error": {
    "field": "email",
    "code": "missing"
  }
}
```

* If the account to delete has a valid subscription, 422 Unprocessable:

```json
{
  "message": "",
  "error": {
    "field": "subscription",
    "code": "already_exists"
  }
}
```

### Response

`204 No Content` if account is deleted succesfully.
