package subscription

import "github.com/guregu/null"

type ConfirmError struct {
	Err   error
	Retry bool
}

func (c ConfirmError) Error() string {
	return c.Err.Error()
}

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

func NewConfirmationResult(orderID string, err error) ConfirmationResult {
	r := ConfirmationResult{
		OrderID:   orderID,
		Succeeded: false,
		Failed:    null.String{},
	}

	if err != nil {
		r.Failed = null.StringFrom(err.Error())
	} else {
		r.Succeeded = true
	}

	return r
}
