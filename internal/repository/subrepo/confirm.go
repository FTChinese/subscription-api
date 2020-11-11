package subrepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

// ConfirmOrder updates the order received from webhook,
// create or update membership, and optionally flag prorated orders as consumed.
func (env Env) ConfirmOrder(result subs.PaymentResult, order subs.Order) (subs.ConfirmationResult, *subs.ConfirmError) {

	defer env.logger.Sync()
	sugar := env.logger.Sugar().With("orderId", result.OrderID)

	sugar.Info("Start confirming order")
	tx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// Step 1: Find the subscription order by order id
	// The row is locked for update.
	// If the order is not found, or is already confirmed,
	// tell provider not sending notification any longer;
	// otherwise, allow retry.
	sugar.Info("Start locking order")
	//order, err := tx.RetrieveOrder(result.OrderID)
	//if err != nil {
	//	sugar.Error(err)
	//	_ = tx.Rollback()
	//	return subs.ConfirmationResult{}, result.ConfirmError(err, err != sql.ErrNoRows)
	//}
	lo, err := tx.LockOrder(result.OrderID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, err != sql.ErrNoRows)
	}

	if !lo.ConfirmedAt.IsZero() {
		_ = tx.Rollback()
		err := errors.New("duplicate confirmation")
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	sugar.Info("Validate payment result")
	if err := order.ValidatePayment(result); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	// STEP 2: query membership
	// For any errors, allow retry.
	sugar.Info("Retrieving existing membership")
	member, err := tx.RetrieveMember(order.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// If user is already a premium member and this order is used
	// for upgrading, decline retry.
	sugar.Info("Validate duplicate upgrading")
	if err := order.ValidateDupUpgrade(member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	sugar.Info("Confirm order")
	confirmed, err := order.Confirm(result, member)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	sugar.Info("Persist confirmed order")
	if err := tx.ConfirmOrder(confirmed.Order); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if member.IsZero() {
		sugar.Infof("Inserting membership %v", confirmed.Membership)
		err := tx.CreateMember(confirmed.Membership)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, result.ConfirmError(err, true)
		}
	} else {
		sugar.Infof("Updating membership %v", confirmed.Membership)
		err := tx.UpdateMember(confirmed.Membership)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, result.ConfirmError(err, true)
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	sugar.Infof("Order %s confirmed", result.OrderID)

	return confirmed, nil
}

func (env Env) SaveConfirmationErr(e *subs.ConfirmError) error {
	_, err := env.db.NamedExec(
		subs.StmtSaveConfirmResult,
		e)

	if err != nil {
		return err
	}

	return nil
}
