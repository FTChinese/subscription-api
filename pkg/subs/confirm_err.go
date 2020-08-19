package subs

import (
	"github.com/guregu/null"
)

type ConfirmError struct {
	Err   error
	Retry bool
}

func (c ConfirmError) Schema(orderID string) ConfirmErrSchema {
	s := ConfirmErrSchema{
		OrderID:   orderID,
		Succeeded: false,
		Failed:    null.String{},
	}

	if c.Err != nil {
		s.Failed = null.StringFrom(c.Err.Error())
	} else {
		s.Succeeded = true
	}

	return s
}

func (c ConfirmError) Error() string {
	return c.Err.Error()
}

// ConfirmErrSchema logs the result of confirmation.
type ConfirmErrSchema struct {
	OrderID   string      `db:"order_id"`
	Succeeded bool        `db:"succeeded"`
	Failed    null.String `db:"failed"`
}
