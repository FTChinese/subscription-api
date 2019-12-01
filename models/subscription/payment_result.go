package subscription

import "time"

// PaymentResult unifies ali and wx webhook notification.
type PaymentResult struct {
	Amount      int64 // Unify various payment amounts to cents.
	OrderID     string
	ConfirmedAt time.Time
}
