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

### Request Body

```json
{
  "ftcId": "uuid of email account"
}
```
### Link Policy

两个账号是否可以关联，主要却决于两个账号下的会员状态。

下列矩阵描述了两个账号会员不同状态下的组合是否允许关联。

| FTC \ Wechat | no member | not expired | expired |
| ------------ | --------- | ----------- | ------- |
| no member    | 1         | 1           | 1       |
| not expired  | 1         | 0           | 1      |
| expired      | 1         | 1           | 1      |

由上可知，只有在双方都有有效会员的情况下会拒绝关联请求。

### Workflow

1. 从头部提取`X-Union-Id`的值；
2. 解析请求的JSON数据。解析错误返回 400 Bad Request;
3. 验证请求的字段。错误返回 422；
4. 分别使用两方ID去除各自的账号数据：使用 `userId` 取出邮箱侧Account，在使用 `unionId` 取出微信侧Account。注意，理想状态下我们应该锁表，但是这里不可以，因为两个账号有可能已经关联，那么取出的数据会是同一样，如果锁表，会导致有一方永远无法取数据。
5. 如果两个账号相等，则认为双方已经关联，返回204；
6. 排除去除了同一个账号后，检测其中是否有任一方已经关联了微信。由于两个账号不同，且至少有一个关联了一个微信，则任一方中必然关联了第三个账号，拒绝关联，返回 422。此处客户端如果想要告诉用户具体哪个账号关联了第三个账号，需要在客户端获取数据判断，返回的错误信息无法提供更多细节。

```json
{
  "message": "one of the accounts or both of them are linked to a 3rd account",
  "error": {
    "field": "account",
    "code": "link_already_taken"
  }
}
```

7. 双方账号都没有关联其他账号，可以执行关联。接下来取决于会员是否可以关联。
8. 首先判断双方取出来的会员是否相等。注意，相等包括空值，即无会员。相等则可以关联。原则上，不应存在账号未关联而会员关联的情况，但由于我们手动操作数据遗留问题，会存在数据不一致的情况
9. 排除相等后，可以确定至少有一方带有会员。如果双方任一方已经关联了第三方账号，则拒绝请求。。返回 422

```json
{
  "message": "at least one of the account's membership is linked to a 3rd account",
  "error": {
    "field": "membership",
    "code": "link_already_taken"
  }
}
```

10. 如果两方会员都没有过期，则拒绝请求，返回422：

```json
{
  "message": "accounts with valid memberships cannot be linked",
  "error": {
    "field": "membership",
    "code": "both_valid"
  }
}
```

11. 会员至少有一方已经过期（空会员属于过期），可以合并，取时间靠后者。
12. 把 union id 保存到 userinfo 表，如果合并后的账号有会员，则删除原有的会员（最多两条记录），插入合并后的会员。被删除的会员记录备份。
13. 发送邮件，告诉用户账号已经绑定。
14. 返回HTTP 200，数据为合并后的Account。

### Example Response

略

## 取消关联

```
POST /account/wx/unlink
```

### Request Body

```json
{
  "ftcId": "string",
  "anchor": "optional enum: ftc | wechat"
}
```

### Workflow

1. 从头部提取`X-Union-Id`的值；
2. 解析请求的JSON数据。解析错误返回 400 Bad Request;
3. 验证请求的字段。错误返回 422；
4. 取出 Account，如果数据库返回错误；
5. 检查 Account 是否关联。如果没有关联，返回 404；
6. 如果取出的 Account unionId 字段和提供的 X-Union-Id 值不相同，则认为请求的账号不存在，返回 404；
7. 验证是否可以断开关联：如果没有会员，则允许；如果有会员，则检查请求中提供的 `anchor` 字段，决定会员保留在断开关联后的哪一方。如果会员来自Stripe、IAP或者B2B，或者账号是测试账号，只能保留在邮箱账号上。
8. 移除useinfo和vip表中对应的字段
9. 发送邮件，告诉用户本次操作细节
10. 返回 204。这里没有返回取消关联后的Account，因为我们不知道用户最初登录采取的哪种方式，因此无法确定到底返回取消关联后的FTC方还是微信方账号。客户端应根据用户取消关联前 Account 中 `loginMethod` 的值刷新 Account。

