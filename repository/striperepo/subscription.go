package striperepo

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

func createSub(p ftcStripe.SubParams, planID string) (*stripe.Subscription, error) {

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(p.Customer),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(planID),
			},
		},
	}

	params.AddExpand("latest_invoice.payment_intent")

	// {
	// "status":400,
	// "message":"Idempotent key length is 0 characters long, which is outside accepted lengths. Idempotent Keys must be 1-255 characters long. If you're looking for a decent generator, try using a UUID defined by IETF RFC 4122.",
	// "request_id":"req_O6zILK5QEVpViw",
	// "type":"idempotency_error"
	// }
	if p.IdempotencyKey != "" {
		logger.Infof("Setting idempotency key: %s", p.IdempotencyKey)
		params.SetIdempotencyKey(p.IdempotencyKey)
	}

	if p.Coupon.Valid {
		params.Coupon = stripe.String(p.DefaultPaymentMethod.String)
	}

	if p.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(p.DefaultPaymentMethod.String)
	}

	return sub.New(params)
}

// CreateSubscription creates a new subscription.
// error could be:
// util.ErrNonStripeValidSub
// util.ErrActiveStripeSub
// util.ErrUnknownSubState
func (env StripeEnv) CreateSubscription(id reader.MemberID, params ftcStripe.SubParams) (*stripe.Subscription, error) {

	ftcPlan, err := params.GetFtcPlan()
	if err != nil {
		return &stripe.Subscription{}, err
	}

	stripePlanID := ftcPlan.GetStripePlanID(env.Live())

	log := logger.WithField("trace", "Stripe.CreateSubscription")

	tx, err := env.beginOrderTx()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Retrieve member for this user to check whether the operation is allowed.
	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		return nil, err
	}

	// Check whether creating stripe subscription is allowed for this member.
	if err := mmb.PermitStripeCreate(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	log.Info("Creating stripe subscription")

	// Contact Stripe API.
	ss, err := createSub(params, stripePlanID)

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
		newMmb := params.NewMembership(id, ss)
		log.Infof("A new stripe membership: %+v", newMmb)
		if err := tx.CreateMember(newMmb); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		newMmb := params.UpdateMembership(mmb, ss)
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
func (env StripeEnv) SaveSubsError(id reader.MemberID, e *stripe.Error) error {
	_, err := env.db.Exec(ftcStripe.StmtSaveStripeError,
		id.FtcID,
		null.NewString(e.ChargeID, e.ChargeID != ""),
		null.NewString(string(e.Code), e.Code != ""),
		null.NewInt(int64(e.HTTPStatusCode), e.HTTPStatusCode != 0),
		e.Msg,
		null.NewString(e.Param, e.Param != ""),
		null.NewString(e.RequestID, e.RequestID != ""),
		e.Type,
	)

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
	if mmb.StripeSubID.IsZero() {
		_ = tx.Rollback()
		return nil, sql.ErrNoRows
	}

	ss, err := sub.Get(mmb.StripeSubID.String, nil)

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
