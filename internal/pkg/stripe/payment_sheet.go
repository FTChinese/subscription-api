package stripe

// PaymentSheet is used by Stripe client-side SDK on Android.
type PaymentSheet struct {
	ClientSecret   string `json:"clientSecret"` // Either payment intent or setup intent client secret
	EphemeralKey   string `json:"ephemeralKey"`
	CustomerID     string `json:"customerId"`
	PublishableKey string `json:"publishableKey"`
	LiveMode       bool   `json:"liveMode"`
}
