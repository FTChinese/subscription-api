package subrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// ConfirmOrder updates the order received from webhook,
// create or update membership, and optionally flag prorated orders as consumed.
// @param order - the order loaded outside a transaction. You have to retrieve and lock it again here.
func (env Env) ConfirmOrder(pr subs.PaymentResult, order subs.Order, p price.Price) (subs.ConfirmationResult, *subs.ConfirmError) {

	defer env.logger.Sync()
	sugar := env.logger.Sugar().
		With("orderId", pr.OrderID).
		With("name", "ConfirmOrder")

	sugar.Info("Start confirming order")
	tx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	// Lock  this order and only retrieves the purchase period.
	sugar.Info("Start locking order")
	lo, err := tx.LockOrder(pr.OrderID)
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), err != sql.ErrNoRows)
	}

	// This step is important to prevent concurrent webhook modification
	// and ensures data integrity.
	order = lo.Merge(order)

	// STEP 2: query membership
	// For any errors, allow retry.
	sugar.Info("Retrieving existing membership")
	member, err := tx.RetrieveMember(order.CompoundID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	sugar.Infof("Existing membership retrieved %v", member)

	// If order is already confirmed, only stop in case it's
	// synced to membership.
	if order.IsConfirmed() {
		sugar.Infof("Duplicate confirmation of order %s", order.ID)
		// If this order is already synced to membership, make no changes.
		if order.IsExpireDateSynced(member) {
			sugar.Infof("Order %s already synced to membership", order.ID)
			_ = tx.Rollback()
			return subs.ConfirmationResult{
				Payment:    pr,
				Order:      order,
				Invoices:   subs.Invoices{},
				Membership: member,
				Snapshot:   reader.MemberSnapshot{},
				Notify:     false,
			}, nil
		}
	}

	// STEP 4: confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	sugar.Info("confirm order")
	confirmed, err := subs.NewConfirmationResult(subs.ConfirmationParams{
		Payment: pr,
		Price:   p,
		Order:   order,
		Member:  member,
	})

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	// Update original order's confirmation time, and optional start
	// time and end time if is not confirmed yet.
	// For any errors, allow retry.
	// NOTE: we are testing the original order, which might be already confirmed in case not synced to membership.
	if !order.IsConfirmed() {
		sugar.Info("Update confirmed order")
		if err := tx.ConfirmOrder(confirmed.Order); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
		}
	}

	sugar.Infof("Invoices generated after confirmation %+v", confirmed.Invoices)
	// Save add-on if having one.
	sugar.Infof("Saving purchase invoice %s", confirmed.Invoices.Purchased.ID)
	err = tx.SaveInvoice(confirmed.Invoices.Purchased)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}
	if !confirmed.Invoices.CarriedOver.IsZero() {
		sugar.Infof("Saving carry-over invoice %s", confirmed.Invoices.CarriedOver.ID)
		err := tx.SaveInvoice(confirmed.Invoices.CarriedOver)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
		}
	}

	// Insert or update membership.
	// This error should allow retry
	// A problem of low possibility discovered
	// when using the policy of updating membership if exists:
	// If current membership is purchased from wechat, the vip_id
	// is union id.
	// Later user linked wechat to FTC  account; however, uuid is not
	// added to vip table for some unknown reason (probably due to manually changing data)
	// Then a new order is created, new membership created with both ids.
	// Since FTC uuid have higher priority, it will be used as the value of vip_id to update this row, which is actually the value of union id!
	if !member.IsZero() {
		sugar.Infof("Deleting old membership %v", member)
		err := tx.DeleteMember(member.UserIDs)
		if err != nil {
			_ = tx.Rollback()
			sugar.Error(err)
		}
	}

	sugar.Infof("Inserting membership %v", confirmed.Membership)
	err = tx.CreateMember(confirmed.Membership)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	sugar.Info("Order confirmation finished")

	return confirmed, nil
}

func (env Env) SaveConfirmErr(e *subs.ConfirmError) error {
	_, err := env.dbs.Write.NamedExec(
		subs.StmtSaveConfirmResult,
		e)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) SavePayResult(result subs.PaymentResult) error {
	if result.OrderID == "" {
		return nil
	}

	_, err := env.dbs.Write.NamedExec(subs.StmtSavePayResult, result)
	if err != nil {
		return err
	}

	return nil
}
