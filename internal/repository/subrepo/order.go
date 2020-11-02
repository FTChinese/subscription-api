package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
)

// For backward compatibility.
const wxAppNativeApp = "wxacddf1c20516eb69" // Used by native app to pay and log in.

func (env Env) CreateOrder(config subs.PaymentConfig) (subs.PaymentIntent, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs.PaymentIntent{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	sugar.Infof("Start retrieving membership for reader %+v", config.Account.MemberID())
	member, err := otx.RetrieveMember(config.Account.MemberID())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.PaymentIntent{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	// Deduce order kind.
	kind, ve := member.AliWxSubsKind(config.Plan.Edition)
	if ve != nil {
		sugar.Error(ve)
		_ = otx.Rollback()
		return subs.PaymentIntent{}, ve
	}
	sugar.Infof("Subscription kind %s", kind)

	// Step 2: Build an order for the user's chosen plan
	// with chosen payment method based on previous
	// membership so that we could how this order
	// is used: create, renew or upgrade.

	var balanceSources []subs.BalanceSource
	// Step 3: required only if this order is used for
	// upgrading.
	if kind == enum.OrderKindUpgrade {
		// Step 3.1: find previous orders with balance
		// remaining.
		// DO not save sources directly. The balance is not
		// calculated at this point.
		sugar.Infof("Get balance sources for an upgrading order")
		balanceSources, err = otx.FindBalanceSources(config.Account.MemberID())
		if err != nil {
			sugar.Error(err)
			_ = otx.Rollback()
			return subs.PaymentIntent{}, err
		}
		sugar.Infof("Find balance source: %+v", balanceSources)
	}

	pi, err := config.BuildIntent(balanceSources, kind)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.PaymentIntent{}, err
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(pi.Order); err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.PaymentIntent{}, err
	}
	sugar.Infof("Order saved %s", pi.Order.ID)

	if err := otx.Commit(); err != nil {
		sugar.Error(err)
		return subs.PaymentIntent{}, err
	}

	return pi, nil
}

func (env Env) SaveProratedOrders(pos []subs.ProratedOrder) error {
	for _, v := range pos {
		_, err := env.db.NamedExec(
			subs.StmtSaveProratedOrder,
			v)

		if err != nil {
			return err
		}
	}

	return nil
}

func (env Env) ProratedOrdersUsed(upOrderID string) error {
	_, err := env.db.Exec(
		subs.StmtProratedOrdersUsed,
		upOrderID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) LogOrderMeta(m subs.OrderMeta) error {

	_, err := env.db.NamedExec(
		subs.StmtInsertOrderMeta,
		m)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveOrder loads an order by its id.
func (env Env) RetrieveOrder(orderID string) (subs.Order, error) {
	var order subs.Order

	err := env.db.Get(
		&order,
		subs.StmtSelectOrder,
		orderID)

	if err != nil {
		return subs.Order{}, err
	}

	// Set wx app id to the one used by native app pay if missing.
	if order.PaymentMethod == enum.PayMethodWx && order.WxAppID.IsZero() {
		order.WxAppID = null.StringFrom(wxAppNativeApp)
	}

	return order, nil
}
