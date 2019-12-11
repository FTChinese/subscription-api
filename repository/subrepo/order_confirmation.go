package subrepo

import (
	"database/sql"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

// ConfirmOrder updates the order received from webhook,
// create or update membership, and optionally flag prorated orders as consumed.
func (env SubEnv) ConfirmOrder(result subscription.PaymentResult) (subscription.Order, *subscription.ConfirmError) {
	log := logger.WithField("trace", "SubEnv.ConfirmOrder")

	tx, err := env.BeginOrderTx()

	if err != nil {
		log.Error(err)
		return subscription.Order{}, &subscription.ConfirmError{
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
		log.Error(err)
		_ = tx.Rollback()
		return subscription.Order{}, &subscription.ConfirmError{
			Err:   err,
			Retry: err != sql.ErrNoRows,
		}
	}

	builder := subscription.
		NewConfirmationBuilder(result, env.Live()).
		SetOrder(order)

	if err := builder.ValidateOrder(); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// STEP 2: query membership
	// For any errors, allow retry.
	member, err := tx.RetrieveMember(order.MemberID)
	if err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, &subscription.ConfirmError{
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
		_ = tx.Rollback()
		return subscription.Order{}, err
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	confirmed, err := builder.Build()
	if err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, &subscription.ConfirmError{
			Err:   err,
			Retry: true,
		}
	}

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	if err := tx.UpdateConfirmedOrder(confirmed.Order); err != nil {
		_ = tx.Rollback()
		return subscription.Order{}, &subscription.ConfirmError{
			Err:   err,
			Retry: false,
		}
	}

	// Flag upgrade balance source as consumed.
	if confirmed.Order.Usage == plan.SubsKindUpgrade {
		err := tx.ProratedOrdersUsed(confirmed.Order.UpgradeSchemaID.String)

		if err != nil {
			_ = tx.Rollback()
			return subscription.Order{}, &subscription.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	}

	// Backup old membership.
	if !confirmed.Snapshot.IsZero() {
		go func() {
			_ = env.BackUpMember(confirmed.Snapshot)
		}()
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if member.IsZero() {
		if err := tx.CreateMember(confirmed.Membership); err != nil {
			_ = tx.Rollback()
			return subscription.Order{}, &subscription.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	} else {
		// Update Membership.
		if err := tx.UpdateMember(confirmed.Membership); err != nil {
			_ = tx.Rollback()
			return subscription.Order{}, &subscription.ConfirmError{
				Err:   err,
				Retry: true,
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return subscription.Order{}, &subscription.ConfirmError{
			Err:   err,
			Retry: true,
		}
	}

	log.Infof("Order %s confirmed", result.OrderID)

	return confirmed.Order, nil
}
