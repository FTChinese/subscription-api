package model

import (
	"github.com/guregu/null"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/paymentintent"
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

func (env Env) CreateStripeSub(
	id paywall.UserID,
	params paywall.StripeSubParams,
) (*stripe.Subscription, error) {

	log := logger.WithField("trace", "Env.CreateStripeCustomer")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return &stripe.Subscription{}, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return &stripe.Subscription{}, nil
	}

	var stripeSub *stripe.Subscription
	switch mmb.ActionOnStripe() {
	default:
		_ = tx.rollback()
		return &stripe.Subscription{}, errors.New("Unknown operation for stripe payment")

	case paywall.StripeActionCreate:
		log.Info("Creating stripe subscription")
		stripeSub, err = createStripeSub(params)

	case paywall.StripeActionSync:
		log.Info("Syncing stripe subscription data")
		stripeSub, err = sub.Get(mmb.StripeSubID.String, nil)

	case paywall.StripeActionNoop:
		log.Info("Stripe noop")
		_ = tx.rollback()
		return &stripe.Subscription{}, ErrAlreadyAMember
	}

	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return &stripe.Subscription{}, err
	}

	newMmb := mmb.FromStripe(id, params, stripeSub)

	if mmb.IsZero() {
		// Insert member
		if err := tx.CreateMember(newMmb, null.StringFrom(params.PlanID)); err != nil {
			_ = tx.rollback()
			return stripeSub, err
		}
	} else {
		// update member
		if err := tx.UpdateMember(newMmb); err != nil {
			_ = tx.rollback()
			return stripeSub, err
		}
	}

	if err := tx.LinkUser(newMmb); err != nil {
		_ = tx.rollback()
		return stripeSub, err
	}

	if err := tx.commit(); err != nil {
		return &stripe.Subscription{}, err
	}

	return &stripe.Subscription{}, nil
}

func (env Env) UpdateStripeSubs(
	id paywall.UserID,
	params paywall.StripeSubParams,
) (*stripe.Subscription, error) {

	log := logger.WithField("trace", "Env.CreateStripeCustomer")

	tx, err := env.BeginOrderTx()
	if err != nil {
		log.Error(err)
		return &stripe.Subscription{}, err
	}

	mmb, err := tx.RetrieveMember(id)
	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return &stripe.Subscription{}, nil
	}

	if !mmb.PermitStripeUpgrade(params) {
		return &stripe.Subscription{}, errors.New("only upgrading from monthly to yearly plan, or from standard product to premium is allowed")
	}

	stripeSub, err := updateStripeSub(params, mmb.StripeSubID.String)

	if err != nil {
		log.Error(err)
		_ = tx.rollback()
		return &stripe.Subscription{}, err
	}

	newMmb := mmb.FromStripe(id, params, stripeSub)

	if err := tx.UpdateMember(newMmb); err != nil {
		_ = tx.rollback()
		return stripeSub, err
	}

	if err := tx.LinkUser(newMmb); err != nil {
		_ = tx.rollback()
		return stripeSub, err
	}

	if err := tx.commit(); err != nil {
		return stripeSub, err
	}

	return stripeSub, nil
}

func createStripeSub(p paywall.StripeSubParams) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(p.Customer),
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

	if p.Coupon.Valid {
		params.Coupon = stripe.String(p.DefaultPaymentMethod.String)
	}

	if p.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(p.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(p.IdempotencyKey)
	return sub.New(params)
}

func updateStripeSub(p paywall.StripeSubParams, subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(p.Customer),
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

	if p.Coupon.Valid {
		params.Coupon = stripe.String(p.DefaultPaymentMethod.String)
	}

	if p.DefaultPaymentMethod.Valid {
		params.DefaultPaymentMethod = stripe.String(p.DefaultPaymentMethod.String)
	}

	params.SetIdempotencyKey(p.IdempotencyKey)
	return sub.Update(subID, params)
}

func createPaymentIntent(price int64, customerID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(price),
		Currency: stripe.String(string(stripe.CurrencyCNY)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Customer: stripe.String(customerID),
	}

	return paymentintent.New(params)
}
