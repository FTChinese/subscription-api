package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/guregu/null"
)

// For backward compatibility.
const wxAppNativeApp = "wxacddf1c20516eb69" // Used by native app to pay and log in.

// CreateOrder creates an order and save it to db.
// This version is applicable to all users, regardless of how their current membership is purchased. It no longer calculates user's
// current account balance.
// For upgrading to premium with valid standard subscription,
// the remaining days is converted to add-on.
// For Stripe and IAP, the purchase is taken as add-on directly.
func (env Env) CreateOrder(counter subs.Counter) (subs.PaymentIntent, error) {
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
	sugar.Infof("Start retrieving membership for reader %+v", counter.BaseAccount.CompoundIDs())
	member, err := otx.RetrieveMember(counter.BaseAccount.CompoundIDs())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs.PaymentIntent{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	pi, err := counter.PaymentIntent(member)
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

func (env Env) SaveOrderMeta(m subs.OrderMeta) error {

	_, err := env.DBs.Write.NamedExec(
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

	err := env.DBs.Read.Get(
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

func (env Env) orderHeader(orderID string) (subs.Order, error) {
	var order subs.Order

	err := env.DBs.Read.Get(
		&order,
		subs.StmtOrderHeader,
		orderID)

	if err != nil {
		return subs.Order{}, nil
	}

	return order, nil
}

func (env Env) orderTail(orderID string) (subs.Order, error) {
	var order subs.Order

	err := env.DBs.Read.Get(
		&order,
		subs.StmtOrderTail,
		orderID)

	if err != nil {
		return subs.Order{}, nil
	}

	return order, nil
}

type orderResult struct {
	order subs.Order
	err   error
}

// LoadFullOrder retrieves an order by splitting a single row into two
// concurrent retrieval since for unknown reasons the DB does not
// respond if a row has two much columns.
func (env Env) LoadFullOrder(orderID string) (subs.Order, error) {
	headerCh := make(chan orderResult)
	tailCh := make(chan orderResult)

	go func() {
		defer close(headerCh)

		order, err := env.orderHeader(orderID)
		headerCh <- orderResult{
			order: order,
			err:   err,
		}
	}()

	go func() {
		defer close(tailCh)

		order, err := env.orderTail(orderID)
		tailCh <- orderResult{
			order: order,
			err:   err,
		}
	}()

	headerRes, tailRes := <-headerCh, <-tailCh
	if headerRes.err != nil {
		return subs.Order{}, headerRes.err
	}

	if tailRes.err != nil {
		return subs.Order{}, tailRes.err
	}

	return subs.Order{
		ID:            headerRes.order.ID,
		UserIDs:       headerRes.order.UserIDs,
		PlanID:        headerRes.order.PlanID,
		DiscountID:    headerRes.order.DiscountID,
		Price:         headerRes.order.Price,
		Edition:       headerRes.order.Edition,
		Charge:        headerRes.order.Charge,
		Kind:          tailRes.order.Kind,
		PaymentMethod: tailRes.order.PaymentMethod,
		WxAppID:       tailRes.order.WxAppID,
		CreatedAt:     tailRes.order.CreatedAt,
		ConfirmedAt:   tailRes.order.ConfirmedAt,
		DatePeriod:    tailRes.order.DatePeriod,
		LiveMode:      true,
	}, nil
}
