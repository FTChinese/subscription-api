## Stripe Setup Intent

The `SetupIntent` object:

```json
{
    "id": "seti_1KUQKcBzTK0hABgJssUtMfl8",
    "cancellationReason": "",
    "clientSecret": "xxx",
    "customerId": "xxx",
    "liveMode": false,
    "nextAction": {},
    "paymentMethodId": null,
    "status": "requires_payment_method",
    "usage": "off_session"
}
```

## Create Setup Intent

```
POST /stripe/setup-intents
```

Create a setup intent so that we could collect user's payment information.

### Request

```json
{
  "customer": "customer id"
}
```

### Response

`SetupIntent`.

### Errors

* `422 Unprocessable` if request body missing


## Get a Setup Intent

```
GET /stripe/setup-intents/{id}
```

### Response

`SetupIntent`.

## Get Setup Intent's Payment Method

```
GET /stripe/setup-intents/{id}/payment-method?<refresh=true>
```

### Response

The `PaymentMethod` object.
