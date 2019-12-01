package subscription

import "github.com/guregu/null"

// ConfirmationResult logs the result of confirmation.
type ConfirmationResult struct {
	OrderID   string      `db:"order_id"`
	Succeeded bool        `db:"succeeded"`
	Failed    null.String `db:"failed"`
	Retry     bool
}

func (r ConfirmationResult) Error() string {
	return r.Failed.String
}

// NewConfirmationSucceeded createa a new instance of ConfirmationResult for success.
func NewConfirmationSucceeded(orderID string) *ConfirmationResult {
	return &ConfirmationResult{
		OrderID:   orderID,
		Succeeded: true,
	}
}

func NewConfirmationFailed(orderID string, reason error, retry bool) *ConfirmationResult {
	return &ConfirmationResult{
		OrderID: orderID,
		Failed:  null.StringFrom(reason.Error()),
		Retry:   retry,
	}
}
