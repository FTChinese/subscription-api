package model

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
)

// CreateStripeCustomer create a customer under ftc account
// for user with `ftcID`.
func (env Env) CreateStripeCustomer(ftcID string) (string, error) {
	tx, err := env.db.Begin()
	if err != nil {
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	var u paywall.FtcUser
	err = env.db.QueryRow(query.LockFtcUser, ftcID).Scan(
		&u.UserID,
		&u.UnionID,
		&u.StripeID,
		&u.UserName,
		&u.Email)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	if u.StripeID.Valid {
		_ = tx.Rollback()
		return u.StripeID.String, nil
	}

	stripeID, err := createCustomer(u.Email)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	_, err = tx.Exec(query.SaveStripeID, stripeID, ftcID)
	if err != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "CreateStripeCustomer").Error(err)
		return "", err
	}

	return stripeID, nil
}

func createCustomer(email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	cus, err := customer.New(params)

	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

// CreateStripeSub creates a new subscription.
func (env Env) CreateStripeSub(
	id paywall.UserID,
	params paywall.StripeSubParams,
) (paywall.StripeSub, error) {

	log := logger.WithField("trace", "Env.CreateStripeSub")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.StripeSub{}, err
	}

	// Retrieve member for this user to check whether the
	// operation is allowed.
	// Only expired members without auto renewal are allowed to create a new subscription via stripe
	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, nil
	}

	if !mmb.PermitStripeCreate() {
		_ = tx.rollback()
		return paywall.StripeSub{}, errors.New("only new members or expired members are allowed to create stripe subscription")
	}

	// When user reaches here, the membership must be expired,
	// and it is not auto renewal.
	//log.Info("Creating stripe subscription")
	s, err := createStripeSub(params)
	// {"status":400,
	// "message":"Keys for idempotent requests can only be used with the same parameters they were first used with. Try using a key other than '4a857eb3-396c-4c91-a8f1-4014868a8437' if you meant to execute a different request.","request_id":"req_Dv6N7d9lF8uDHJ",
	// "type":"idempotency_error"
	// }
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, err
	}

	//log.Infof("Payment intent %+v", stripeSub.LatestInvoice.PaymentIntent)
	newMmb, err := mmb.FromStripe(id, s)
	if err != nil {
		return paywall.StripeSub{}, err
	}

	log.Infof("A stripe membership: %+v", newMmb)

	if mmb.IsZero() {
		// Insert member
		if err := tx.CreateMember(newMmb); err != nil {
			_ = tx.rollback()
			return paywall.StripeSub{}, err
		}
	} else {
		// update member
		if err := tx.UpdateMember(newMmb); err != nil {
			_ = tx.rollback()
			return paywall.StripeSub{}, err
		}
	}

	if err := tx.commit(); err != nil {
		return paywall.StripeSub{}, err
	}

	return paywall.NewStripeSub(s), nil
}

// GetStripeSub refresh stripe subscription data if stale.
func (env Env) GetStripeSub(id paywall.UserID) (paywall.StripeSub, error) {
	log := logger.WithField("trace", "Env.GetStripeSub")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.StripeSub{}, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, nil
	}

	if mmb.IsZero() {
		return paywall.StripeSub{}, sql.ErrNoRows
	}

	log.Infof("Retrieve a member: %+v", mmb)

	log.Info("Getting stripe subscription")

	s, err := sub.Get(mmb.StripeSubID.String, &stripe.SubscriptionParams{
		Params: stripe.Params{
			Expand: []*string{
				stripe.String("latest_invoice.payment_intent"),
			},
		},
	})

	if err != nil {
		//log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, err
	}

	//log.Infof("Payment intent %+v", stripeSub.LatestInvoice.PaymentIntent)
	newMmb, err := mmb.FromStripe(id, s)
	if err != nil {
		return paywall.StripeSub{}, err
	}

	log.Infof("Refreshed membership: %+v", newMmb)

	//log.Infof("A stripe membership: %+v", newMmb)
	// update member
	if err := tx.UpdateMember(newMmb); err != nil {
		_ = tx.rollback()
		return paywall.StripeSub{}, err
	}

	if err := tx.commit(); err != nil {
		return paywall.StripeSub{}, err
	}

	log.Info("Refreshed finished.")
	return paywall.NewStripeSub(s), nil
}

// UpgradeStripeSubs switches subscription plan.
func (env Env) UpgradeStripeSubs(
	id paywall.UserID,
	params paywall.StripeSubParams,
) (paywall.StripeSub, error) {

	log := logger.WithField("trace", "Env.CreateStripeCustomer")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return paywall.StripeSub{}, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, nil
	}

	if !mmb.PermitStripeUpgrade(params) {
		_ = tx.rollback()
		return paywall.StripeSub{}, errors.New("only upgrading from standard member to premium is allowed")
	}

	s, err := upgradeStripeSub(params, mmb.StripeSubID.String)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return paywall.StripeSub{}, err
	}

	newMmb, err := mmb.FromStripe(id, s)
	if err != nil {
		return paywall.StripeSub{}, err
	}

	if err := tx.UpdateMember(newMmb); err != nil {
		_ = tx.rollback()
		return paywall.StripeSub{}, err
	}

	if err := tx.commit(); err != nil {
		return paywall.StripeSub{}, err
	}

	return paywall.NewStripeSub(s), nil
}

func createStripeSub(p paywall.StripeSubParams) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(p.Customer),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(p.PlanID),
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

func upgradeStripeSub(p paywall.StripeSubParams, subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(p.PlanID),
			},
		},
		Params: stripe.Params{
			Expand: []*string{
				stripe.String("latest_invoice.payment_intent"),
			},
		},
	}

	if p.IdempotencyKey != "" {
		params.IdempotencyKey = stripe.String(p.IdempotencyKey)
	}

	if p.Coupon.Valid {
		params.Coupon = stripe.String(p.DefaultPaymentMethod.String)
	}

	if p.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(p.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(p.IdempotencyKey)
	return sub.Update(subID, params)
}
