package subrepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
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
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: true,
		}
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
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: err != sql.ErrNoRows,
		}
	}

	builder := subs.NewConfirmationBuilder(result).
		SetOrder(order)

	// Validate order
	if err := builder.ValidateOrder(); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	// STEP 2: query membership
	// For any errors, allow retry.
	member, err := tx.RetrieveMember(order.MemberID)
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: true,
		}
	}

	builder.SetMembership(member)

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// If user is already a premium member and this order is used
	// for upgrading, decline retry.
	if err := builder.ValidateDuplicateUpgrading(); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, err
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	confirmed, err := builder.Build()
	if err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: true,
		}
	}

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	if err := tx.ConfirmOrder(confirmed.Order); err != nil {
		sugar.Error(err)
		_ = tx.Rollback()
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: false,
		}
	}

	// Flag upgrade balance source as consumed.
	if confirmed.Order.Kind == enum.OrderKindUpgrade {
		err := tx.ProratedOrdersUsed(confirmed.Order.ID)

		if err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, &subs.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if member.IsZero() {
		if err := tx.CreateMember(confirmed.Membership); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, &subs.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	} else {
		// Update Membership.
		if err := tx.UpdateMember(confirmed.Membership); err != nil {
			sugar.Error(err)
			_ = tx.Rollback()
			return subs.ConfirmationResult{}, &subs.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	}

	if err := tx.Commit(); err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, &subs.ConfirmError{
			Err:   err,
			Retry: true,
		}
	}

	sugar.Infof("Order %s confirmed", result.OrderID)

	return confirmed, nil
}

func (env Env) SaveConfirmationResult(r subs.ConfirmErrSchema) error {
	_, err := env.db.NamedExec(
		subs.StmtSaveConfirmResult,
		r)

	if err != nil {
		return err
	}

	return nil
}
