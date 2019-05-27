# Gift Card

## Redeem

    PUT /gift-card/redeem
    
### Headers

One of the following, or both if user's ftc account is bound to wechat account:

```
X-User-Id: xxxx
X-Union-Id: xxx
```

### Input

```json
{
  "code": "redeem code"
}
```

### Response

* `400 Bad Request` there's any error parsing JSON.

* `422 Unprocessable Entity` 

if `code` is missing from request body:

```json
{
  "message": "",
  "error": {
    "field": "redeem_code",
    "code": "missing_field"
  }
}
```

if this user is already a member:

```json
{
  "message": "",
  "error": {
    "field": "member",
    "code": "already_exists"
  }
}
```

* `204 No Content` if the gift code is valid and membership is created for this user.

After the card is used, its code is marked as being used. The action and the membership creation are performed in a transaction, an all-or-nothing operation to ensure data integrity.