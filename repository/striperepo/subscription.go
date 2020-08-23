package striperepo

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
)

// CreateSubscription creates a new subscription.
// error could be:
// util.ErrNonStripeValidSub
// util.ErrActiveStripeSub
// util.ErrUnknownSubState
func (env StripeEnv) CreateSubscription(input ftcStripe.SubsInput) (*stripe.Subscription, error) {

	log := logger.WithField("trace", "Stripe.CreateSubscription")

	tx, err := env.beginOrderTx()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(input.FtcID),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	subsKind, ve := mmb.StripeSubsKind(input.Edition)
	if ve != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if subsKind != enum.OrderKindCreate {
		_ = tx.Rollback()
		return nil, errors.New("invalid request to create Stripe subscription")
	}

	log.Info("Creating stripe subscription")

	// Contact Stripe API.
	ss, err := input.CreateSubs()

	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	log.Infof("A stripe membership: %+v", mmb)

	// What if the user exists but is invalid?
	if mmb.IsZero() {
		newMmb := input.NewMembership(ss)
		log.Infof("A new stripe membership: %+v", newMmb)
		if err := tx.CreateMember(newMmb); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		newMmb := input.UpdateMembership(mmb, ss)
		log.Infof("Updated stripe membership: %+v", newMmb)
		if err := tx.UpdateMember(newMmb); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ss, err
}

// SaveSubsError saves any error in stripe response.
func (env StripeEnv) SaveSubsError(e ftcStripe.APIError) error {
	_, err := env.db.NamedExec(ftcStripe.StmtSaveAPIError, e)

	if err != nil {
		logger.WithField("trace", "SubEnv.SaveSubsError").Error(err)
		return err
	}

	return nil
}

// GetSubscription refresh stripe subscription data if stale.
func (env StripeEnv) GetSubscription(id reader.MemberID) (*stripe.Subscription, error) {
	log := logger.WithField("trace", "StripeEvn.GetSubscription")

	tx, err := env.beginOrderTx()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	if mmb.IsZero() {
		log.Infof("Membership not found for %v", id)
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	log.Infof("Retrieve a member: %+v", mmb)

	// If this membership is not a stripe subscription, deny further actions
	if mmb.PaymentMethod != enum.PayMethodStripe {
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}
	if mmb.StripeSubsID.IsZero() {
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	ss, err := ftcStripe.GetSubscription(mmb.StripeSubsID.String)

	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	newMmb := ftcStripe.RefreshMembership(mmb, ss)

	log.Infof("Refreshed membership: %+v", newMmb)

	// update member
	if err := tx.UpdateMember(newMmb); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Info("Refreshed finished.")
	return ss, nil
}
