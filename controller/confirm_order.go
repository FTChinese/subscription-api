package controller

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

func (router PayRouter) confirmPayment(result paywall.PaymentResult) (paywall.Order, *paywall.ConfirmationResult) {
	log := logrus.WithField("trace", "PayRouter.confirmPayment")

	tx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
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

		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, err != sql.ErrNoRows)
	}

	log.Infof("Order to confirm %s", order.ID)

	if order.IsConfirmed() {
		log.Infof("Order %s is already confirmed. Abort.", order.ID)

		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, util.ErrAlreadyConfirmed, false)
	}

	if order.AmountInCent() != result.Amount {
		log.Infof("Paid amount does not match. Expected %f, actual %d", order.Amount, result.Amount)
		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, fmt.Errorf("amount mismatched: expected: %d, actual: %d", order.AmountInCent(), result.Amount), false)
	}

	log.Infof("Found order %s", order.ID)

	// STEP 2: query membership
	// For any errors, allow retry.
	member, err := tx.RetrieveMember(order.GetAccountID())
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// If user is already a premium member and this order is used
	// for upgrading, decline retry.
	if order.Usage == paywall.SubsKindUpgrade && member.IsValidPremium() {
		log.Infof("Order %s is trying to upgrade a premium member %s", order.ID, member.ID.String)
		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, util.ErrDuplicateUpgrading, false)
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	confirmedOrder, err := order.Confirm(member, result.ConfirmedAt)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}
	log.Infof("Order %s confirmed : %s - %s", result.OrderID, order.StartDate, order.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	if err := tx.ConfirmOrder(confirmedOrder); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	newMember, err := member.FromAliOrWx(order)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()

		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if member.IsZero() {
		if err := tx.CreateMember(newMember); err != nil {
			log.Error(err)
			_ = tx.Rollback()
			return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
		}
	} else {
		if err := tx.UpdateMember(newMember); err != nil {
			log.Error(err)
			_ = tx.Rollback()
			return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return paywall.Order{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	log.Infof("Order %s confirmed", result.OrderID)

	return confirmedOrder, nil
}
