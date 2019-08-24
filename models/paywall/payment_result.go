package paywall

import "time"

// PaymentResult unifies ali and wx webhook notification.
type PaymentResult struct {
	Amount      int64
	OrderID     string
	ConfirmedAt time.Time
}
