package paywall

import (
	"encoding/json"
	"github.com/stripe/stripe-go"
	"testing"
)

const invoiceData = `
{
    "id": "in_1ExnStBzTK0hABgJO5BJw0Bc",
    "object": "invoice",
    "account_country": "GB",
    "account_name": "THE FINANCIAL TIMES LIMITED",
    "amount_due": 3000,
    "amount_paid": 3000,
    "amount_remaining": 0,
    "application_fee_amount": null,
    "attempt_count": 1,
    "attempted": true,
    "auto_advance": false,
    "billing": "charge_automatically",
    "billing_reason": "subscription_create",
    "charge": "ch_1ExnczBzTK0hABgJDLbC6Sfu",
    "collection_method": "charge_automatically",
    "created": 1563509582,
    "currency": "gbp",
    "custom_fields": null,
    "customer": "cus_FOgRRgj9aMzpAv",
    "customer_address": null,
    "customer_email": "neefrankie@gmail.com",
    "customer_name": null,
    "customer_phone": null,
    "customer_shipping": null,
    "customer_tax_exempt": "none",
    "customer_tax_ids": [],
    "default_payment_method": null,
    "default_source": null,
    "default_tax_rates": [],
    "description": null,
    "discount": null,
    "due_date": null,
    "ending_balance": 0,
    "footer": null,
    "hosted_invoice_url": "https://pay.stripe.com/invoice/invst_4bPSPelrhZkV1P0exGhFFZqJTU",
    "invoice_pdf": "https://pay.stripe.com/invoice/invst_4bPSPelrhZkV1P0exGhFFZqJTU/pdf",
    "lines": {
        "object": "list",
        "data": [
            {
                "id": "sli_64abb8ed4726c7",
                "object": "line_item",
                "amount": 3000,
                "currency": "gbp",
                "description": "1 × Standard (at £30.00 / year)",
                "discountable": true,
                "livemode": false,
                "metadata": {},
                "period": {
                    "end": 1595131982,
                    "start": 1563509582
                },
                "plan": {
                    "id": "plan_FOdfeaqzczp6Ag",
                    "object": "plan",
                    "active": true,
                    "aggregate_usage": null,
                    "amount": 3000,
                    "billing_scheme": "per_unit",
                    "created": 1562567504,
                    "currency": "gbp",
                    "interval": "year",
                    "interval_count": 1,
                    "livemode": false,
                    "metadata": {},
                    "nickname": "Standard Yearly Plan",
                    "product": "prod_FOde1wE4ZTRMcD",
                    "tiers": null,
                    "tiers_mode": null,
                    "transform_usage": null,
                    "trial_period_days": null,
                    "usage_type": "licensed"
                },
                "proration": false,
                "quantity": 1,
                "subscription": "sub_FSiuLINrqxCiDt",
                "subscription_item": "si_FSiubsF9m72b5M",
                "tax_amounts": [],
                "tax_rates": [],
                "type": "subscription"
            }
        ],
        "has_more": false,
        "total_count": 1,
        "url": "/v1/invoices/in_1ExnStBzTK0hABgJO5BJw0Bc/lines"
    },
    "livemode": false,
    "metadata": {},
    "next_payment_attempt": null,
    "number": "6C9D489E-0016",
    "paid": true,
    "payment_intent": "pi_1ExnStBzTK0hABgJnGlZKt8Q",
    "period_end": 1563509582,
    "period_start": 1563509582,
    "post_payment_credit_notes_amount": 0,
    "pre_payment_credit_notes_amount": 0,
    "receipt_number": null,
    "starting_balance": 0,
    "statement_descriptor": null,
    "status": "paid",
    "status_transitions": {
        "finalized_at": 1563509583,
        "marked_uncollectible_at": null,
        "paid_at": 1563510210,
        "voided_at": null
    },
    "subscription": "sub_FSiuLINrqxCiDt",
    "subtotal": 3000,
    "tax": null,
    "tax_percent": null,
    "total": 3000,
    "total_tax_amounts": [],
    "webhooks_delivered_at": null
}`

func TestNewStripeInvoice(t *testing.T) {
	var i stripe.Invoice

	if err := json.Unmarshal([]byte(invoiceData), &i); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", i)

	t.Logf("Payment intent: %+v", i.PaymentIntent)

	t.Logf("Customer: %+v", i.Customer)
}
