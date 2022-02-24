## The Customer object

Response of `Customer`
```json
{
  "id": "cus_KXMeDzH46HnpMB",
  "ftcId": "4eab2991-9669-41c2-b51e-75e1a0e76183",
  "currency": "gbp",
  "created": 1636078372,
  "defaultSource": null,
  "defaultPaymentMethod": null,
  "email": "xxxxxx",
  "liveMode": false
}
```

## Create Stripe Customer

```
POST /stripe/customers
```

在使用Stripe订阅之前，必须先在Stripe[创建 Customer](https://stripe.com/docs/api/customers/create).

### Request

NULL

### Response

`Customer`

### Errors

See Stripe's error description.

## Load a stripe customer

```
GET /stripe/customers/{id}?<refresh=true>
```

* `id`: stripe customer id.
* `refresh`: if provided, always fetch data from Stripe API rather than using our DB.

### Response

`Customer`

### Errors

* `404 Not Found` is customer is not found, or the provided ftc id and customer id does not match.

## Get customer default payment method

```
GET /stripe/customers/{id}/default-payment-method?<refresh=true>
```

* `id: string`: customer id
* `refresh: boolean` Optional.

Usually used to display to user the payment method used when creating a new subscription.

### Response

See Payment Method object

### Errors

* 404 Not Found if the customer does not have default payment set.

## Set/Change Customer Default Payment Method

```
POST /stripe/customers/{id}/default-payment-method
```

* `id`: customer id

### Request Body

```json
{
  "defaultPaymentMethod": "id of a payment method"
}
```

### Response

`Customer`.

### Errors

* `400 Bad Request` if request body cannot be parsed
* `422 Unprocessable` if request body is invalid

```json
{
  "message": "Missing required field",
  "error": {
    "field": "defaultPaymentMethod",
    "code": "missing_field"
  }
}
```

## Customer Payment Methods

```
GET /stripe/customers/{id}/default-payment-method?<refresh=true>
```

Retrieve a customer's default payment method.

### Response

`PaymentMethod`.

## List Customer Payment Methods

```
GET /stripe/customers/{id}/payment-methods
```

List all payment methods belonging to a customer

### Response

```json
{
  "total": 3,
  "page": 1,
  "limit": 20,
  "data": []
}
```

`data` is an array of `PaymentMethod`.

## 获取Ephemeral Keys

```
POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
```

URL参数 `api_version` 为必填项，从客户端SDK中获取。

Stripe API的数据原样返回，客户端SDK直接使用。
