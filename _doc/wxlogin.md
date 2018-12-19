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

## Load a Wechat User's Account

    GET /user/account

Request **Must** contain the following header:

```
X-Union-Id: "the unionId you get from the last step"
```

### Response

* `404 Not Found` if the `unionId` is not found in database.

* `200 OK`

```json
{
    "id": "ftc uuid",
    "userName": "string | null",
    "email": "login email",
    "avatarUrl": null,
    "isVip": false,
    "isVerified": false,
    "wechat": {
        "unionID": "",
        "openID": "",
        "nickName": "",
        "avatarUrl"
    },
    "membership": {
        "tier": "standard | premium | or empty for free user",
        "billingCycle": "month | year | or empty",
        "expireDate": "2019-12-12"
    }
}
```

`wechat` field could be `null`, but here it should always has value since it is retrieved by wechat union id. If you request user account data to `next-api`, which retrieve user account by FTC's user id, `wechat` could be `null`.

If `id` is empty string, it indicates this wechat account is not bound to any FTC account.

## Bind Accounts

    PUT /user/bind

Bind a wechat account to FTC account.

Request header must contain `X-Union-Id`.

Client should ask user to validate ftc account credentials to make sure the user actually owns the ftc account.

Account binding could be split into two parts: binding accounts for `cmstmp01.userinfo` table and binding memberships for `premium.ftc_vip` table. Binding accounts is simple: just add `wx_union_id` column if it is null. Bindig membeships involves merging membership, deleting one, inserting or updating one and save the deleted one to another table for backup.

### Input

```json
{
    "userId": "ftc user id"
}
```

### Response

* `204 No Content` if two accounts are bound.

* `400 Bad Request` if request body cannot be pased.

```json
{
    "message": "Problems parsing JSON"
}
```

* `422 Unprocessable Entity` if `userId` is empty.

```json
{
    "message": "Validation failed",
    "error": {
        "field": "userId",
        "code": "missing_field"
    }
}
```

* `403 Forbidden`

if any of FTC account of Wechat account is bound to a 3rd account:
```json
{
    "message": "One of the requested accounts, or both, is/are bound to a 3rd account"
}
```

if the membership of any of the two accounts are bound to a 3rd account:
```json
{
    "message": "The membership of one of the requested accounts, or both, is/are bound to a 3rd account"
}
```

if the membership of both accounts are not bound to a 3rd accounts, but they are not expired yet:
```json
{
    "message": "The two accounts have different valid memberships!"
}
```

### 绑定策略

注意：客户端向此处发起请求前，应该告诉用户当前两个账号的会员状态，在客户端计算出来是否存在会员信息合并的情况，让用户确认以后再请求API执行下属操作。客户端的计算方式可以参考此处列出的步骤。

1. 用`userId`和`unionId`为条件分别取出这两个id所对应的账号。

用`userId`取数据所采用的SQL：
```sql
SELECT u.user_id AS id,
    u.user_name AS userName,
    u.email AS email,
-- 会员信息
-- LEFT JOIN 左侧列必然存在，右侧列则都有可能为NULL，
-- 与数据库定义时是否允许为NULL无关。
-- v.vip_id 为空则表明该用户不是会员。
    IFNULL(v.vip_id, ''),
    v.vip_id_alias,
    IFNULL(v.expire_date, '') AS expireDate,
-- 微信账号ID
-- 空表明该FTC账号没有绑定微信账号。
    IFNULL(unionid, '') AS unionId,
    IFNULL(openid, '') AS openId,
    IFNULL(nickname, '') AS nickName,
    IFNULL(headimgurl, '') AS avatarUrl
FROM cmstmp01.userinfo AS u
    LEFT JOIN premium.ftc_vip AS v
    ON u.user_id = v.vip_id
    LEFT JOIN user_db.user_sns_info AS w
    ON u.wx_union_id = w.unionid
WHERE u.user_id = ?
LIMIT 1
```
我们把这个账号称为**FTC账号**。

用`unionId`取出同样数据，但是JOIN方向和WHERE条件不同：
```sql
-- 以存储微信账号的表为核心
SELECT w.unionid AS unionId,
-- 会员信息，可能为空
    IFNULL(v.vip_id, ''),
    v.vip_id_alias,
    IFNULL(v.expire_date, '') AS expireDate,
-- FTC账号，也可能为空
    IFNULL(u.user_id, '') AS id,
    IFNULL(u.email, '') AS email,
FROM user_db.user_sns_info AS w
    LEFT JOIN premium.ftc_vip AS v
    ON w.unionid = v.vip_id_alias
    LEFT JOIN cmstmp01.userinfo AS u
    ON w.unionid = u.wx_union_id
WHERE w.unionid = ?
LIMIT 1`
```
我们把这个账号称为**微信账号**。

2. 比较两个账号是否相等，比较条件是这两个账号的`userId`是否相等。因为`useinfo`表中的`user_id`列是唯一的，相等则意味着，这两种SQL语句取出了同样的数据。也就是说，两个账号已经绑定了。结束请求。

3. 如果不相等，分别检测两个账号是否绑定了其他账号：对FTC账号而言，`wechat`字段不是`null`；对微信账号而言，`userId`字段不为空。如果两个账号有任意一个检测出绑定了其他账号，则拒绝绑定请求。

4. 此时已经可以确定，FTC账号没有绑定微信账号，微信账号也没有绑定FTC账号，两个账号是可以绑定的。此时的两个账号是独立存在的。检查两个账号是否都购买了会员。

5. 如果两个账号都没有购买会员，只需要绑定账号即可，就是在`userinfo`表中，找到`userId`行，把`wx_union_id`设为请求中的微信`unionId`。请求结束。

6. 如果两个账号都购买了会员，则检测两个账号取出来的会员信息是否是同一个，判断条件是会员信息的`userId`是否相等。如果会员信息是同一个，则表明这是此前的遗留问题：客服为用户绑定了会员信息，但是账号是没有绑定的，按第五步中的操作绑定账号即可。

7. 两个账号都有会员，但是会员信息不是同一个，检测两个会员信息中是否有绑定另外的账号，如果其中任意一个绑定了一个其他的账号，则拒绝请求。

8. 两个账号都有会员，会员信息不是同一个，且两个账号的会员信心都没有绑定任何其他账号，则表明两个账号可以合并，检测两个账号的到期时间，如果两个账号都没有到期，则拒绝请求。

9. 如果任意一个账号的会员到期了，或者都到期了，而可以合并会员，两个会员信息中所有字段取最大：会员类型`standard`和`premium`中保留`premium`，周期中`year`和`month`保留`year`，`expireDate`中取靠后的日期。然后，把微信账号取出的会员信息保存到`merged_member`表中留存；删除`ftc_vip`表中用微信购买的会员（如果存在），插入一条新纪录，新纪录保存合并的会员信息，新的会员记录会同时有FTC账号的user id和微信账号的union id。这里的插入语句采用了`ON DUPLICATE KEY UPDATE`来处理user id已经存在的情况。一切以FTC账号体系为核心，减少不确定性。

10. 至此可以确定，两个账号中只有一个购买了会员，挑出购买了会员的那个账号的会员信息，执行上一步中相同的删除、创建会员，以及绑定账号。

这里绑定账号、绑定会员使用了SQL的transaction，确保两处的绑定操作同时成功或失败。