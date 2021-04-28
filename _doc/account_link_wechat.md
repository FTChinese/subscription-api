# 关联微信和邮箱账号

## Endpoint

发送到此处的URL请求均需带有Header `X-Union-Id: <wechat union id>`，否则请求会被拒绝。

* GET `/account/wx` 获取微信登录用户的`Account`
* POST `/account/wx/signup` 微信登录用户创建并关联新的邮箱账号
* POST `/account/wx/link` 微信登录用户关联已有邮箱账号，或者邮箱登录用户关联微信
* POST `/account/wx/unlink` 断开已经关联的账号

## 获取微信登录用户账号数据

```
GET /account/wx
```

### Workflow

1. 从头部提取`X-Union-Id`的值；
2. `user_db.wechat_userinifo` LEFT JOIN `cmstmp01.userinfo`
3. 找到数据返回200
4. 未找到数据返回404

### Example Response

```json
{
    "id": "",
    "unionId": "5XLXc9Wzd4YM6-Vk2pj6Wh3MXqYw",
    "stripeId": null,
    "email": "",
    "mobile": null,
    "userName": null,
    "avatarUrl": null,
    "isVerified": false,
    "loginMethod": "wechat",
    "wechat": {
        "nickname": "6Cunningham",
        "avatarUrl": "https://randomuser.me/api/portraits/thumb/men/17.jpg"
    },
    "membership": {
        "ftcId": null,
        "unionId": "5XLXc9Wzd4YM6-Vk2pj6Wh3MXqYw",
        "tier": "standard",
        "cycle": "year",
        "expireDate": "2020-07-25",
        "payMethod": "wechat",
        "stripeSubsId": null,
        "autoRenew": false,
        "status": null,
        "appleSubsId": null,
        "b2bLicenceId": null,
        "standardAddOn": 0,
        "premiumAddOn": 0,
        "vip": false
    }
}
```

## 微信用户注册新邮箱账号

```
POST /account/wx/signup
```

### Request

```json
{
  "email": "string",
  "password": "string",
  "sourceUrl": "url used to verify email, optional."
}
```

`sourceUrl`用以确定邮箱注册时发送的验证邮件中的验证地址，默认为`next-user`运行的地址。如果`next-user`运行在多个域名下，该字段可以提供更多灵活性。

### Workflow

1. 从头部提取`X-Union-Id`的值；
2. 解析请求的JSON数据。解析错误返回 400 Bad Request;
3. 验证请求的字段。错误返回 422；
4. 检测邮箱是否已经存在。存在返回 422；
5. 使用微信 union id 取出微信侧账号，账号数据同 `/account/wx`，同新建的账号合并。
6. 保存新账号；
7. 如果合并后的账号没有会员（即当前微信账号没有会员）,结束；
8. 否则合并后的会员同时存在ftc的uuid和微信的union id，为防止mysql unique id问题，先删除现有会员，再保存合并后的会员。
9. 并发进程任务：备份被删除的会员信息，如果存在；并保存用户的ip等元数据；发送验证邮件
10. 返回 `Account`

### Example Response

```json
{
    "id": "77efa3d0-f2c4-44a6-95b7-d2d82f3f250e",
    "unionId": "5XLXc9Wzd4YM6-Vk2pj6Wh3MXqYw",
    "stripeId": null,
    "email": "aliwx.test@ftchinese.com",
    "mobile": null,
    "userName": "Test",
    "avatarUrl": null,
    "isVerified": false,
    "loginMethod": "email",
    "wechat": {
          "nickname": "6Cunningham",
          "avatarUrl": "https://randomuser.me/api/portraits/thumb/men/17.jpg"
    },
    "membership": {
        "ftcId": "77efa3d0-f2c4-44a6-95b7-d2d82f3f250e",
        "unionId": "5XLXc9Wzd4YM6-Vk2pj6Wh3MXqYw",
        "tier": "standard",
        "cycle": "year",
        "expireDate": "2022-01-19",
        "payMethod": "alipay",
        "stripeSubsId": null,
        "autoRenew": false,
        "status": null,
        "appleSubsId": null,
        "b2bLicenceId": null,
        "standardAddOn": 0,
        "premiumAddOn": 0,
        "vip": false
    }
}
```

### 关联现有账号

```
POST /account/wx/link
```

### Request

```json
{
  "userId": "uuid of email account"
}
```

### Workflow

1. 从头部提取`X-Union-Id`的值；
2. 解析请求的JSON数据。解析错误返回 400 Bad Request;
3. 验证请求的字段。错误返回 422；

### Example Response
