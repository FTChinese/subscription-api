package subscription

import "github.com/guregu/null"

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

// FreeUpgrade contains the data after creating/upgrading/renewing a membership.
// All data here is in a final state.
// Those data can be directly save into database.
// The generated of those fields has a intertwined dependency
// on each other, so they are return in one batch.
type ConfirmationResult struct {
	Order      Order          // The confirmed order.
	Membership Membership     // The updated membership.
	Snapshot   MemberSnapshot // // Snapshot of previous membership
}
