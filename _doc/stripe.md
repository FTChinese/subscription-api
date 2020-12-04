# Stripe Subscription

To enable subscription, you have to first [Create a customer](https://stripe.com/docs/api/customers/create) at Stripe.

## Status

* `incomplete`
* `incomplete_expired`
* `trialing`
* `active`
* `past_due`
* `canceled`
* `unpaid`

```
                 [not paid in 23 hours]
     incomplete -----------------------> incomplete_expired 
    |         \ 
  [failed]     \ [paid]
    |           \               /--> past_due
(initial pay)    | ---> active | ---> canceled
                /               \--> unpaid
               /
  trialing    /
```

## Cancel

There are 3 approaches to cancel a subscription:

* Cancel immediately. 

This uses the cancel endpoint `DELETE /v1/subscription/:id`. The `canceled_at` field will be populated with a timestamp which is the date of this cancellation.

See [Cancel a subscription](https://stripe.com/docs/api/subscriptions/cancel).

* Cancel at period end. 

This uses the update endpoint `POST /v1/subscription/:id`, supplying parameter `cancel_at_period_end`. The `canceled_at` field will also be populated with a timestamp. However, it will reflect the most recent update request rather than the cancellation time.

See [Update a subscription](https://stripe.com/docs/api/subscriptions/update).

* Cancel at a specified time.

This also uses the update endpoint, supplying parameter `cancel_at`. It could be either a past or future time.

We only use *cancel at period end* approach.

### Determine expiration date

Use `status`, `canceld_at` together with `cancel_at_period_end` to determine membership's expiration date and auto renewal:

```
var autoRenew: bool
var expirationDate: Date

if status == "canceled" {
    autoRenew = false

    if !cancel_at && !cancel_at_period_end {
        expirationDate = canceled_at
        return        
    }
    
    if cancel_at_period_end {
        expirationDate = current_period_end
        return
    }

    expirationDate = cancel_at
}
```

## Subscription Kind

Since all subscriptions are auto renewable by default, a subscription could be either creation or upgrading. There won't be renewal cases.




