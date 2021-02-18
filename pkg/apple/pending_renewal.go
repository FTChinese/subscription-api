package apple

import (
	"github.com/guregu/null"
	"strconv"
)

// PendingRenewal contains auto-renewable subscription renewals that are open or failed in the past.
type PendingRenewal struct {
	// only present if the user downgrades or crossgrades to a subscription of a different duration for the subsequent subscription period
	AutoRenewProductID null.String `json:"auto_renew_product_id"`
	// The renewal status for the auto-renewable subscription.
	// 1: The subscription will renew at the end of the current subscription period.
	// 0: The customer has turned off automatic renewal for the subscription.
	// The value for this field should not be interpreted as the customer’s subscription status
	AutoRenewStatus string `json:"auto_renew_status"`
	// The reason a subscription expired.
	// This field is only present for a receipt that contains an expired auto-renewable subscription.
	// 1 The customer voluntarily canceled their subscription.
	// 2 Billing error; for example, the customer's payment information was no longer valid.
	// 3 The customer did not agree to a recent price increase.
	// 4 The price was not available for purchase at the time of renewal.
	// 5 Unknown error.
	ExpirationIntent null.String `json:"expiration_intent"`
	// The time at which the grace period for subscription renewals expires
	GracePeriodExpiresDate string `json:"grace_period_expires_date"`
	// This key is only present for apps that have Billing Grace Period enabled and when the user experiences a billing error at the time of renewal.
	GracePeriodExpiresDateMs  null.String `json:"grace_period_expires_date_ms"`
	GracePeriodExpiresDatePST string      `json:"grace_period_expires_date_pst"`
	// A flag that indicates Apple is attempting to renew an expired subscription automatically.
	// This field is only present if an auto-renewable subscription is in the billing retry state
	// 1 - The App Store is attempting to renew the subscription.
	// 0 - The App Store has stopped attempting to renew the subscription.
	IsInBillingRetryPeriod null.String `json:"is_in_billing_retry_period"`
	// The transaction identifier of the original purchase.
	OriginalTransactionID string `json:"original_transaction_id"`
	// This field is only present if the customer was notified of the price increase.
	// The default value is "0" and changes to "1" if the customer consents.
	PriceConsentStatus null.String `json:"price_consent_status"`
	ProductID          string      `json:"product_id"`
}

func (p PendingRenewal) IsAutoRenew() bool {
	ok, err := strconv.ParseBool(p.AutoRenewStatus)
	if err != nil {
		return false
	}

	return ok
}

func (p PendingRenewal) Schema(e Environment) PendingRenewalSchema {
	return PendingRenewalSchema{
		BaseSchema: BaseSchema{
			Environment:           e,
			OriginalTransactionID: p.OriginalTransactionID,
		},
		ProductID:          p.ProductID,
		AutoRenewStatus:    p.AutoRenewStatus,
		ExpirationIntent:   p.ExpirationIntent,
		AutoRenewProductID: p.AutoRenewProductID,
		IsInBillingRetryPeriod: null.NewBool(
			MustParseBoolean(p.IsInBillingRetryPeriod.String),
			p.IsInBillingRetryPeriod.Valid),
		GracePeriodExpiresDateMs: null.NewInt(
			MustParseInt64(p.GracePeriodExpiresDateMs.String),
			p.GracePeriodExpiresDateMs.Valid),
		PriceConsentStatus: p.PriceConsentStatus,
	}
}
