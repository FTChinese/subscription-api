package paywall

import (
	"encoding/json"
	"fmt"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
	"time"
)

var gender = []string{"men", "women"}

func GenWxID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func GenAvatar() string {
	n := randomdata.Number(1, 35)
	g := gender[randomdata.Number(0, 2)]

	return fmt.Sprintf("https://randomuser.me/api/portraits/thumb/%s/%d.jpg", g, n)
}

func GenSubID() string {
	id, _ := gorest.RandomBase64(9)
	return "sub_" + id
}

type AccountKind int

const (
	AccountKindFtc AccountKind = iota
	AccountKindWx
	AccountKindLinked
)

type Profile struct {
	FtcID      string
	UnionID    string
	Email      string
	Password   string
	UserName   string
	Avatar     string
	OpenID     string
	ExpireDate chrono.Date
	IP         string
}

func NewProfile() Profile {
	return Profile{
		FtcID:      uuid.New().String(),
		UnionID:    GenWxID(),
		Email:      fake.EmailAddress(),
		Password:   fake.SimplePassword(),
		UserName:   fake.UserName(),
		Avatar:     GenAvatar(),
		OpenID:     GenWxID(),
		ExpireDate: chrono.DateNow(),
		IP:         fake.IPv4(),
	}
}

func (p Profile) UserID(kind AccountKind) UserID {
	var id UserID

	switch kind {
	case AccountKindFtc:
		id, _ = NewID(p.FtcID, "")

	case AccountKindWx:
		id, _ = NewID("", p.UnionID)

	case AccountKindLinked:
		id, _ = NewID(p.FtcID, p.UnionID)
	}

	return id
}

func (p Profile) RandomKindUserID() UserID {
	return p.UserID(AccountKind(randomdata.Number(0, 3)))
}

func (p Profile) FtcUser() FtcUser {
	return FtcUser{
		UserID:   p.FtcID,
		UnionID:  null.String{},
		StripeID: null.String{},
		Email:    p.Email,
		UserName: null.StringFrom(p.UserName),
	}
}

func (p Profile) Membership(kind AccountKind, pm enum.PayMethod, expired bool) Membership {
	m := NewMember(p.UserID(kind))
	m.Tier = standardYearlyPlan.Tier
	m.Cycle = standardYearlyPlan.Cycle

	if expired {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, 0, -7))
	} else {
		m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	}

	m.PaymentMethod = pm

	if pm == enum.PayMethodStripe {
		m.StripeSubID = null.StringFrom(GenSubID())
		m.StripePlanID = null.StringFrom(stripeTestPlanIDs["standard_year"])
		m.AutoRenewal = true
		m.Status = SubStatusActive
	}

	return m
}

func (p Profile) AliWxSub(kind AccountKind, pm enum.PayMethod, usage SubsKind) Subscription {
	s, err := NewSubs(p.UserID(kind), standardYearlyPlan)
	if err != nil {
		panic(err)
	}

	s.ConfirmedAt = chrono.TimeNow()
	s.IsConfirmed = true
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))
	s.PaymentMethod = pm
	s.StartDate = chrono.DateNow()
	s.Usage = usage
	if pm == enum.PayMethodWx {
		s.WxAppID = null.StringFrom(GetWxAppID())
	}

	return s
}

func GetStripeSub() stripe.Subscription {
	s := stripe.Subscription{}
	if err := json.Unmarshal([]byte(subData), &s); err != nil {
		panic(err)
	}

	return s
}

func GetWxAppID() string {
	return viper.GetString("wxapp.m_subs.app_id")
}

var subData = `
{
  "id": "sub_FRie8eetfXxfpW",
  "object": "subscription",
  "application_fee_percent": null,
  "billing": "charge_automatically",
  "billing_cycle_anchor": 1563277950,
  "billing_thresholds": null,
  "cancel_at": null,
  "cancel_at_period_end": false,
  "canceled_at": null,
  "collection_method": "charge_automatically",
  "created": 1563277950,
  "current_period_end": 1594900350,
  "current_period_start": 1563277950,
  "customer": "cus_FOgRRgj9aMzpAv",
  "days_until_due": null,
  "default_payment_method": "pm_1Ett5HBzTK0hABgJwXpA8b7z",
  "default_source": null,
  "default_tax_rates": [],
  "discount": null,
  "ended_at": null,
  "items": {
    "object": "list",
    "data": [
      {
        "id": "si_FRie5twOfaFJ9r",
        "object": "subscription_item",
        "billing_thresholds": null,
        "created": 1563277951,
        "metadata": {},
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
        "quantity": 1,
        "subscription": "sub_FRie8eetfXxfpW",
        "tax_rates": []
      }
    ],
    "has_more": false,
    "total_count": 1,
    "url": "/v1/subscription_items?subscription=sub_FRie8eetfXxfpW"
  },
  "latest_invoice": {
    "id": "in_1EwpCsBzTK0hABgJMsWJpxtG",
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
    "charge": "ch_1EwpCuBzTK0hABgJLe3dd2az",
    "collection_method": "charge_automatically",
    "created": 1563277950,
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
    "hosted_invoice_url": "https://pay.stripe.com/invoice/invst_7w4r7nLr0fylK5nMaG7nVEE6Zs",
    "invoice_pdf": "https://pay.stripe.com/invoice/invst_7w4r7nLr0fylK5nMaG7nVEE6Zs/pdf",
    "lines": {
      "object": "list",
      "data": [
        {
          "id": "sli_1bbae456dcabd8",
          "object": "line_item",
          "amount": 3000,
          "currency": "gbp",
          "description": "1 × Standard (at £30.00 / year)",
          "discountable": true,
          "livemode": false,
          "metadata": {},
          "period": {
            "end": 1594900350,
            "start": 1563277950
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
          "subscription": "sub_FRie8eetfXxfpW",
          "subscription_item": "si_FRie5twOfaFJ9r",
          "tax_amounts": [],
          "tax_rates": [],
          "type": "subscription"
        }
      ],
      "has_more": false,
      "total_count": 1,
      "url": "/v1/invoices/in_1EwpCsBzTK0hABgJMsWJpxtG/lines"
    },
    "livemode": false,
    "metadata": {},
    "next_payment_attempt": null,
    "number": "6C9D489E-0007",
    "paid": true,
    "payment_intent": {
      "id": "pi_1EwpCsBzTK0hABgJWF2EOnbd",
      "object": "payment_intent",
      "amount": 3000,
      "amount_capturable": 0,
      "amount_received": 3000,
      "application": null,
      "application_fee_amount": null,
      "canceled_at": null,
      "cancellation_reason": null,
      "capture_method": "automatic",
      "charges": {
        "object": "list",
        "data": [
          {
            "id": "ch_1EwpCuBzTK0hABgJLe3dd2az",
            "object": "charge",
            "amount": 3000,
            "amount_refunded": 0,
            "application": null,
            "application_fee": null,
            "application_fee_amount": null,
            "balance_transaction": "txn_1EwpCuBzTK0hABgJvdq88Y4r",
            "billing_details": {
              "address": {
                "city": null,
                "country": null,
                "line1": null,
                "line2": null,
                "postal_code": null,
                "state": null
              },
              "email": null,
              "name": null,
              "phone": null
            },
            "captured": true,
            "created": 1563277952,
            "currency": "gbp",
            "customer": "cus_FOgRRgj9aMzpAv",
            "description": "Payment for invoice 6C9D489E-0007",
            "destination": null,
            "dispute": null,
            "failure_code": null,
            "failure_message": null,
            "fraud_details": {},
            "invoice": "in_1EwpCsBzTK0hABgJMsWJpxtG",
            "livemode": false,
            "metadata": {},
            "on_behalf_of": null,
            "order": null,
            "outcome": {
              "network_status": "approved_by_network",
              "reason": null,
              "risk_level": "normal",
              "risk_score": 42,
              "seller_message": "Payment complete.",
              "type": "authorized"
            },
            "paid": true,
            "payment_intent": "pi_1EwpCsBzTK0hABgJWF2EOnbd",
            "payment_method": "pm_1Ett5HBzTK0hABgJwXpA8b7z",
            "payment_method_details": {
              "card": {
                "brand": "visa",
                "checks": {
                  "address_line1_check": null,
                  "address_postal_code_check": null,
                  "cvc_check": null
                },
                "country": "US",
                "exp_month": 8,
                "exp_year": 2020,
                "fingerprint": "nL3BDVQutZ3lff1S",
                "funding": "credit",
                "last4": "4242",
                "three_d_secure": null,
                "wallet": null
              },
              "type": "card"
            },
            "receipt_email": null,
            "receipt_number": null,
            "receipt_url": "https://pay.stripe.com/receipts/acct_1EpW3EBzTK0hABgJ/ch_1EwpCuBzTK0hABgJLe3dd2az/rcpt_FRiecehA6gm4DcOtAkyCBhQXuhC19ay",
            "refunded": false,
            "refunds": {
              "object": "list",
              "data": [],
              "has_more": false,
              "total_count": 0,
              "url": "/v1/charges/ch_1EwpCuBzTK0hABgJLe3dd2az/refunds"
            },
            "review": null,
            "shipping": null,
            "source": null,
            "source_transfer": null,
            "statement_descriptor": null,
            "status": "succeeded",
            "transfer_data": null,
            "transfer_group": null
          }
        ],
        "has_more": false,
        "total_count": 1,
        "url": "/v1/charges?payment_intent=pi_1EwpCsBzTK0hABgJWF2EOnbd"
      },
      "client_secret": "pi_1EwpCsBzTK0hABgJWF2EOnbd_secret_OmJNq8nSigylnFBBnCGMoNjqs",
      "confirmation_method": "automatic",
      "created": 1563277950,
      "currency": "gbp",
      "customer": "cus_FOgRRgj9aMzpAv",
      "description": "Payment for invoice 6C9D489E-0007",
      "invoice": "in_1EwpCsBzTK0hABgJMsWJpxtG",
      "last_payment_error": null,
      "livemode": false,
      "metadata": {},
      "next_action": null,
      "on_behalf_of": null,
      "payment_method": "pm_1Ett5HBzTK0hABgJwXpA8b7z",
      "payment_method_options": {
        "card": {
          "request_three_d_secure": "automatic"
        }
      },
      "payment_method_types": [
        "card"
      ],
      "receipt_email": null,
      "review": null,
      "setup_future_usage": null,
      "shipping": null,
      "source": null,
      "statement_descriptor": null,
      "status": "succeeded",
      "transfer_data": null,
      "transfer_group": null
    },
    "period_end": 1563277950,
    "period_start": 1563277950,
    "post_payment_credit_notes_amount": 0,
    "pre_payment_credit_notes_amount": 0,
    "receipt_number": null,
    "starting_balance": 0,
    "statement_descriptor": null,
    "status": "paid",
    "status_transitions": {
      "finalized_at": 1563277950,
      "marked_uncollectible_at": null,
      "paid_at": 1563277952,
      "voided_at": null
    },
    "subscription": "sub_FRie8eetfXxfpW",
    "subtotal": 3000,
    "tax": null,
    "tax_percent": null,
    "total": 3000,
    "total_tax_amounts": [],
    "webhooks_delivered_at": null
  },
  "livemode": false,
  "metadata": {},
  "pending_setup_intent": null,
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
  "quantity": 1,
  "schedule": null,
  "start": 1563277950,
  "start_date": 1563277950,
  "status": "active",
  "tax_percent": null,
  "trial_end": null,
  "trial_start": null
}`
