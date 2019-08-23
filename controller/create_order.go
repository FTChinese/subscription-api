package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

// createOrder creates an order for ali or wx pay.
func (router PayRouter) createOrder(
	id paywall.AccountID,
	plan paywall.Plan,
	method enum.PayMethod,
	app util.ClientApp,
	wxAppId null.String,
) (paywall.Subscription, error) {
	log := logrus.WithField("trace", "PayRouter.createOrder")

	if method != enum.PayMethodWx && method != enum.PayMethodAli {
		return paywall.Subscription{}, errors.New("only used by alipay or wxpay")
	}

	otx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	log.Infof("Start retrieving membership for reader %+v", id)
	member, err := otx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}
	log.Infof("Membership retrieved %+v", member)

	// Optional: add member id is this member exists but
	// its id field is empty.
	if !member.IsZero() && member.ID.IsZero() {
		member.GenerateID()

		go func() {
			if err := router.env.AddMemberID(member); err != nil {
				log.Error(err)
			}
		}()
	}

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.
	order, err := paywall.NewOrder(id, plan, method, member)
	if err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}

	// Step 3: required only if this order is used for
	// upgrading.
	if order.Usage == paywall.SubsKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		sources, err := otx.FindBalanceSources(id)
		if err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}
		log.Infof("Find balance source: %+v", sources)

		// Step 3.2: Build upgrade plan
		up := paywall.NewUpgradePreview(sources)

		// Step 3.3: Update order based on upgrade plan.
		order = order.WithUpgrade(up)

		// Step 3.4: Save the upgrade plan
		if err := otx.SaveUpgradePlan(up); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}

		// Step 3.5: Save prorated orders
		if err := otx.SaveProration(up.Data); err != nil {
			log.Error(err)
			_ = otx.Rollback()
			return paywall.Subscription{}, err
		}
	}

	snapshot := paywall.NewMemberSnapshot(member)
	order.WxAppID = wxAppId
	order.MemberSnapshotID = null.StringFrom(snapshot.ID)

	// Step 4: Save this order.
	if err := otx.SaveOrder(order); err != nil {
		log.Error(err)
		_ = otx.Rollback()
		return paywall.Subscription{}, err
	}

	if err := otx.Commit(); err != nil {
		log.Error(err)
		return paywall.Subscription{}, err
	}

	// These two steps are not vital.
	// Perform in background.
	go func() {
		if err := router.env.SaveOrderClient(order.ID, app); err != nil {
			log.Error(err)
		}
	}()

	// Back up membership state the moment the order is created.
	go func() {
		if err := router.env.BackUpMember(snapshot); err != nil {
			log.Error(err)
		}
	}()

	if !router.env.Live() {
		order.Amount = 0.01
	}

	return order, nil
}

func (router PayRouter) confirmPayment(result paywall.PaymentResult) (paywall.Subscription, *paywall.ConfirmationResult) {
	log := logrus.WithField("trace", "PayRouter.confirmPayment")

	tx, err := router.env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
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

		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, err != sql.ErrNoRows)
	}

	if order.IsConfirmed() {
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, util.ErrAlreadyConfirmed, false)
	}

	if order.AmountInCent() != result.Amount {
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, fmt.Errorf("amount mismatched: expected: %d, actual: %d", order.AmountInCent(), result.Amount), false)
	}

	log.Infof("Found order %s", order.ID)

	// STEP 2: query membership
	// For any errors, allow retry.
	member, err := tx.RetrieveMember(order.GetAccountID())
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	// STEP 3: validate the retrieved order.
	// This order might be invalid for upgrading.
	// If user is already a premium member and this order is used
	// for upgrading, decline retry.
	if order.Usage == paywall.SubsKindUpgrade && member.IsValidPremium() {
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, util.ErrDuplicateUpgrading, false)
	}

	// STEP 4: Confirm this order
	// Populate the ConfirmedAt, StartDate and EndDate.
	// If there are calculation errors, allow retry.
	confirmedOrder, err := order.Confirm(member, result.ConfirmedAt)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}
	log.Infof("Order %s confirmed : %s - %s", result.OrderID, order.StartDate, order.EndDate)

	// STEP 5: Update confirmed order
	// For any errors, allow retry.
	if err := tx.ConfirmOrder(confirmedOrder); err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	newMember, err := member.FromAliOrWx(order)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()

		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	// STEP 7: Insert or update membership.
	// This error should allow retry
	if member.IsZero() {
		if err := tx.CreateMember(newMember); err != nil {
			log.Error(err)
			_ = tx.Rollback()
			return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
		}
	} else {
		if err := tx.UpdateMember(newMember); err != nil {
			log.Error(err)
			_ = tx.Rollback()
			return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return paywall.Subscription{}, paywall.NewConfirmationFailed(result.OrderID, err, true)
	}

	log.Infof("Order %s confirmed", result.OrderID)

	return confirmedOrder, nil
}
