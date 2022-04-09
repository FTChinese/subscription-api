package txrepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
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
func (tx OrderTx) SaveOrder(order ftcpay.Order) error {

	_, err := tx.NamedExec(
		ftcpay.StmtCreateOrder,
		order)

	if err != nil {
		return err
	}

	return nil
}

func (tx OrderTx) LockOrder(orderID string) (ftcpay.LockedOrder, error) {
	var lo ftcpay.LockedOrder

	err := tx.Get(&lo, ftcpay.StmtLockOrder, orderID)

	if err != nil {
		return ftcpay.LockedOrder{}, err
	}

	return lo, nil
}

// ConfirmOrder set an order's confirmation time and the purchased period.
func (tx OrderTx) ConfirmOrder(order ftcpay.Order) error {
	_, err := tx.NamedExec(
		ftcpay.StmtConfirmOrder,
		order,
	)

	if err != nil {
		return err
	}

	return nil
}
