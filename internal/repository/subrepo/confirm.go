package subrepo

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// ConfirmOrder updates the order received from webhook,
// create or update membership, and optionally flag prorated orders as consumed.
func (env Env) ConfirmOrder(pr subs.PaymentResult, order subs.Order) (subs.ConfirmationResult, *subs.ConfirmError) {

	defer env.logger.Sync()
	sugar := env.logger.Sugar().With("orderId", pr.OrderID).With("name", "ConfirmOrder")

	sugar.Info("Start confirming order")
	tx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	// Step 1: Find the order by order id and lock it.
	sugar.Info("Start locking order")
	_, err = tx.LockOrder(pr.OrderID)
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), err != sql.ErrNoRows)
	}

	// STEP 2: query membership
	// For any errors, allow retry.
	sugar.Info("Retrieving existing membership")
	member, err := tx.RetrieveMember(order.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	sugar.Infof("Existing membership retrieved %v", member)

	// Change nothing.
	if order.IsSynced(member) {
		_ = tx.Rollback()
		return subs.ConfirmationResult{
			Payment: pr,
			PaymentConfirmed: subs.PaymentConfirmed{
				Order:    order,
				AddOn:    subs.AddOn{},
				Snapshot: reader.MemberSnapshot{},
			},
			Notify: false,
		}, nil
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	sugar.Info("Confirm order")
	confirmed, err := subs.NewConfirmationResult(subs.ConfirmationParams{
		Payment: pr,
		Order:   order,
		Member:  member,
	})

	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
	}

	// Update confirmed order if is not confirmed yet.
	// For any errors, allow retry.
	if !order.IsConfirmed() {
		sugar.Info("Update confirmed order")
		if err := tx.ConfirmOrder(confirmed.Order); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, pr.ConfirmError(err.Error(), true)
		}
	}

	// Save add-on if having one.
	if !confirmed.AddOn.IsZero() {
		sugar.Info("Creating add-on")
		if err := tx.SaveAddOn(confirmed.AddOn); err != nil {
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
		err := tx.DeleteMember(member.MemberID)
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
	_, err := env.rwdDB.NamedExec(
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

	_, err := env.rwdDB.NamedExec(subs.StmtSavePayResult, result)
	if err != nil {
		return err
	}

	return nil
}
