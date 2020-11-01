package subrepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"go.uber.org/zap"
)

// ConfirmOrder updates the order received from webhook,
// create or update membership, and optionally flag prorated orders as consumed.
func (env Env) ConfirmOrder(result subs.PaymentResult) (subs.ConfirmationResult, *subs.ConfirmError) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

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
	order, err := tx.RetrieveOrder(result.OrderID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, err != sql.ErrNoRows)
	}

	if order.IsConfirmed() {
		_ = tx.Rollback()
		err := errors.New("duplicate confirmation")
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	if err := order.ValidatePayment(result); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	// STEP 2: query membership
	// For any errors, allow retry.
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
	if err := order.ValidateDupUpgrade(member); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, false)
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	confirmed, err := order.Confirm(result, member)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	if err := tx.ConfirmOrder(confirmed.Order); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if !member.IsZero() {
		err := tx.DeleteMember(member.MemberID)
		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, result.ConfirmError(err, true)
		}
	}

	err = tx.CreateMember(confirmed.Membership)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, result.ConfirmError(err, true)
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
