## The Payment Method object

```json
{
    "id": "xxx",
    "customerId": "xxx",
    "kind": "card",
    "card": {
        "brand": "mastercard",
        "country": "US",
        "expMonth": 12,
        "expYear": 2024,
        "fingerprint": "xxx",
        "funding": "credit",
        "last4": "4444",
        "network": {
            "available": [
                "mastercard"
            ],
            "preferred": ""
        },
        "threeDSecureUsage": {
            "supported": true
        }
    },
    "created": 1645178646,
    "liveMode": false
}
```

## Load Payment Method

```
GET /stripe/payment-methods/{id}?<refresh=true>
```

Load a payment method by its id.

### Response

The `PaymentMethod` object
