package txrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
)

// OrderTx check a user's member status and create an order
// if allowed.
type OrderTx struct {
	SharedTx
}

func NewOrderTx(tx *sqlx.Tx) OrderTx {
	return OrderTx{
		SharedTx: NewSharedTx(tx),
	}
}

// SaveOrder saves an order to db.
// This is only limited to alipay and wechat pay.
// Stripe pay does not generate any orders on our side.
func (tx OrderTx) SaveOrder(order subs.Order) error {

	_, err := tx.NamedExec(
		subs.StmtInsertOrder,
		order)

	if err != nil {
		return err
	}

	return nil
}

func (tx OrderTx) LockOrder(orderID string) (subs.LockedOrder, error) {
	var lo subs.LockedOrder

	err := tx.Get(&lo, subs.StmtLockOrder, orderID)

	if err != nil {
		return subs.LockedOrder{}, err
	}

	return lo, nil
}

// ConfirmOrder set an order's confirmation time and the purchased period.
func (tx OrderTx) ConfirmOrder(order subs.Order) error {
	_, err := tx.NamedExec(
		subs.StmtConfirmOrder,
		order,
	)

	if err != nil {
		return err
	}

	return nil
}
