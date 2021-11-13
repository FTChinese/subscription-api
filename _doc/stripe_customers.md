## 创建Stripe用户

在使用Stripe订阅之前，必须先在Stripe[创建 Customer](https://stripe.com/docs/api/customers/create).

```
POST /stripe/customers
```

### Workflow

1. 从HTTP Header中获取UUID。

2. 锁表

3. 从数据库取出用户账号。未找到返回 `404 Not Found`

4. 如果数据中 `stripe_customer_id` 字段存在，则认为该账号已经注册了Stripe用户，使用该值从Stripe获取到用户数据，返回，结束。

5. 如果尚未注册Stripe用户，调用Stripe SDK创建用户，提交的数据中仅包含用户的邮箱。

6. 保存 Customer ID 到 `stripe_customer_id` 字段。

遇到的任何数据库错误返回 `500 Internval Server Error`

### Example Response

```json
{
    "id": "cus_IXp31Fk2jYJmU3",
    "ftcId": "c07f79dc-664b-44ca-87ea-42958e7991b0",
    "defaultSource": null,
    "defaultPaymentMethod": "pm_1Hzzx3BzTK0hABgJGy155ZR1",
    "email": "stripe.test@ftchinese.com",
    "liveMode": false,
    "createdUtc": "2020-12-10T07:17:54Z"
}
```


## 获取Stripe Customer的详情

```
GET /stripe/customers/{id}
```

ID是Stripe customer的id。改请求仍需提供FTC的UUID，用于验证二者是否一致。

### Workflow

1. 从HTTP header获取UUID。

2. 获取URL中的 `id` 值。

3. 使用UUID获取用户账号数据

4. 检查 `stripe_customer_id` 是否为空。如果是，则该 Stripe customer 不存在，返回404。

5. 检查 `stripe_customer_id` 是否与 `id` 相同，不相同返回404。此举旨在防止客户端发送的数据错误导致FTC用户A获取到了FTC用户B的Stripe数据。

6. 从Stripe API获取customer数据。错误和返回数据同上一节。

## 更新默认支付方式

```
POST /stripe/customers/{id}/default-payment-method
```

### Request Body

```json
{
  "defaultPaymentMethod": "id of a payment method"
}
```

### Workflow

1. 从header获取uuid，url获取customer id，request body获取请求参数。

2. 如果request body无法解析，返回 `400 Bad Request`；如果数据不合法，返回 `422 Unprocessable`:

```json
{
  "message": "Missing required field",
  "error": {
    "field": "defaultPaymentMethod",
    "code": "missing_field"
  }
}
```

3. 使用uuid取出用户账号。未找到、账号不存在stripe customer id，或不等于url中的id，返回404。

4. 发送请求到Stripe。

5. 返回数据同上一节。

## 获取Ephemeral Keys

```
POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
```

URL参数 `api_version` 为必填项，从客户端SDK中获取。

Stripe API的数据原样返回，客户端SDK直接使用。
