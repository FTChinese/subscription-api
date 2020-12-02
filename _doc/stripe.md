# Stripe Subscription

To enable subscription, you have to first [Create a customer](https://stripe.com/docs/api/customers/create) at Stripe.

## Cancel

There are two approaches to cancel a subscription:

* Cancel immediately. This uses the cancel endpoint `DELETE /v1/subscription/:id`. See [Cancel a subscription](https://stripe.com/docs/api/subscriptions/cancel)
* Cancel at period end. This uses the update ednpoint `POST /v1/subscription/:id`, supplying parameter `cancel_at_period_end`. See [Update a subscription](https://stripe.com/docs/api/subscriptions/update). 


