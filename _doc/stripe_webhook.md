## Stripe Webhook

```
POST /webhook/stripe
```

There's only one webhook endpoint for various notifications, which are distinguished by the `type` field in payload.

### Subscription

Used to handle 3 event types:

* `customer.subscription.created`
* `customer.subscripiton.updated`
* `customer.subscripiton.deleted`

Workflow as follows:

1. JSON parse the payaload's `Data.Raw` field.

2. Send a signal back to Stripe immediatly. We'll process the result in background.

3. Then retrieve user account from  user table by `stripe_customer_id` column so that we could be sure this is a Stripe user.

4. In case of any error other than not found, stop. If the error is not found, we continue to insert/update this Stripe subscription record into `stripe_subscription` table and stop. User's membership won't be touched.

5. Start to process subscription data by locking tables.

6. We first try to retrieve membership from `ftc_vip` table by `stripe_subscription_id`. If the result if not found in such a way, we will try to retrieve membership from `ftc_vip` table by uuid or wechat id. Mind here that you must retrieve both side in a mutually exclusive manner; otherwise you might cause a problem of locking the same row twice.

7. Next we start build a new membership based on:

    * the Stripe subscription
    * user account
    * Stripe-side membership
    * Ftc-side membership.
    


8. You have to handle the following possibilities:

9. No membership corresponds to stripe subscription id. In such case what you can do depends on the ftc side:

    * Ftc side has no membership. You are safe to create a stripe membership directly;

    * Ftc side has membership but expired. You can override it.

    * Ftc side is a one-time purchase. :

        * Stripe subscription expired, you are not allowed to touch ftc membership

        * Stripe subscription is not expired. This indicates a one-time purchase is switching to Stripe. Override ftc side with carry over addon.

10. Stripe side has membership:

    * If user id from Stripe-side membership does not match the account retrieved using customer id, it indicates data inconsistency, stop;

    * Otherwise the stripe membership already exists, simply update it.

11. Save/Update stripe subscription into `stripe_subscripiton` table.

12. Save memberships prior and after change to `member_version` table for inspection.
