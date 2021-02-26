// +build !production

package faker

import (
	"encoding/json"
	"github.com/stripe/stripe-go/v72"
)

const StripeSubs = `
{
    "id": "sub_IX3JAkik1JKDzW",
    "object": "subscription",
    "application_fee_percent": null,
    "billing_cycle_anchor": 1607407057,
    "billing_thresholds": null,
    "cancel_at": null,
    "cancel_at_period_end": false,
    "canceled_at": null,
    "collection_method": "charge_automatically",
    "created": 1607407057,
    "current_period_end": 1638943057,
    "current_period_start": 1607407057,
    "customer": "cus_FRgIy7R6sn5nI7",
    "days_until_due": null,
    "default_payment_method": null,
    "default_source": null,
    "default_tax_rates": [],
    "discount": null,
    "ended_at": null,
    "items": {
        "object": "list",
        "data": [
            {
                "id": "si_IX3JJ7rrB8wmwY",
                "object": "subscription_item",
                "billing_thresholds": null,
                "created": 1607407058,
                "metadata": {},
                "plan": {
                    "id": "price_1IM2nFBzTK0hABgJiIDeDIox",
                    "object": "plan",
                    "active": true,
                    "aggregate_usage": null,
                    "amount": 23800,
                    "amount_decimal": "23800",
                    "billing_scheme": "per_unit",
                    "created": 1562567431,
                    "currency": "gbp",
                    "interval": "year",
                    "interval_count": 1,
                    "livemode": false,
                    "metadata": {
                        "tier": "premium",
                        "cycle": "year"
                    },
                    "nickname": "Premium Yearly Price",
                    "price": "prod_FOdd1iNT29BIGq",
                    "tiers_mode": null,
                    "transform_usage": null,
                    "trial_period_days": null,
                    "usage_type": "licensed"
                },
                "price": {
                    "id": "price_1IM2nFBzTK0hABgJiIDeDIox",
                    "object": "price",
                    "active": true,
                    "billing_scheme": "per_unit",
                    "created": 1562567431,
                    "currency": "gbp",
                    "livemode": false,
                    "lookup_key": null,
                    "metadata": {
                        "tier": "premium",
                        "cycle": "year"
                    },
                    "nickname": "Premium Yearly Price",
                    "price": "prod_FOdd1iNT29BIGq",
                    "recurring": {
                        "aggregate_usage": null,
                        "interval": "year",
                        "interval_count": 1,
                        "trial_period_days": null,
                        "usage_type": "licensed"
                    },
                    "tiers_mode": null,
                    "transform_quantity": null,
                    "type": "recurring",
                    "unit_amount": 23800,
                    "unit_amount_decimal": "23800"
                },
                "quantity": 1,
                "subscription": "sub_IX3JAkik1JKDzW",
                "tax_rates": []
            }
        ],
        "has_more": false,
        "total_count": 1,
        "url": "/v1/subscription_items?subscription=sub_IX3JAkik1JKDzW"
    },
    "latest_invoice": "in_1HvzCfBzTK0hABgJklW9Azi5",
    "livemode": false,
    "metadata": {},
    "next_pending_invoice_item_invoice": null,
    "pause_collection": null,
    "pending_invoice_item_interval": null,
    "pending_setup_intent": null,
    "pending_update": null,
    "plan": {
        "id": "plan_FOde0uAr0V4WmT",
        "object": "plan",
        "active": true,
        "aggregate_usage": null,
        "amount": 23800,
        "amount_decimal": "23800",
        "billing_scheme": "per_unit",
        "created": 1562567431,
        "currency": "gbp",
        "interval": "year",
        "interval_count": 1,
        "livemode": false,
        "metadata": {
            "tier": "premium",
            "cycle": "year"
        },
        "nickname": "Premium Yearly Price",
        "price": "prod_FOdd1iNT29BIGq",
        "tiers_mode": null,
        "transform_usage": null,
        "trial_period_days": null,
        "usage_type": "licensed"
    },
    "quantity": 1,
    "schedule": null,
    "start_date": 1607407057,
    "status": "active",
    "transfer_data": null,
    "trial_end": null,
    "trial_start": null
}`

func GenStripeSubs() (*stripe.Subscription, error) {
	var ss stripe.Subscription
	if err := json.Unmarshal([]byte(StripeSubs), &ss); err != nil {
		return nil, err
	}

	return &ss, nil
}

func MustGenStripeSubs() *stripe.Subscription {
	sub, err := GenStripeSubs()
	if err != nil {
		panic(err)
	}

	return sub
}
