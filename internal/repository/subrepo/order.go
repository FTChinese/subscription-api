package subrepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	subs2 "github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

// For backward compatibility.
const wxAppNativeApp = "***REMOVED***" // Used by native app to pay and log in.

// CreateOrder creates an order and save it to db.
// This version is applicable to all users, regardless of how their current membership is purchased. It no longer calculates user's
// current account balance.
// For upgrading to premium with valid standard subscription,
// the remaining days is converted to add-on.
// For Stripe and IAP, the purchase is taken as add-on directly.
func (env Env) CreateOrder(counter subs2.Counter) (subs2.PaymentIntent, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return subs2.PaymentIntent{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	sugar.Infof("Start retrieving membership for reader %+v", counter.BaseAccount.CompoundIDs())
	member, err := otx.RetrieveMember(counter.BaseAccount.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs2.PaymentIntent{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	// TODO: avoid using a discount multiple times
	// First calculate a Checkout instance.
	// Then see if the Offer field has Recurring.
	// If not recurring, retrieve from db the usage history
	// by user id and offer id.
	// If not found, continue to calculate payment intent;
	// if found,
	pi, err := counter.PaymentIntent(member)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs2.PaymentIntent{}, err
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(pi.Order); err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return subs2.PaymentIntent{}, err
	}
	sugar.Infof("Order saved %s", pi.Order.ID)

	if err := otx.Commit(); err != nil {
		sugar.Error(err)
		return subs2.PaymentIntent{}, err
	}

	return pi, nil
}

func (env Env) SaveOrderMeta(c footprint.OrderClient) error {

	_, err := env.dbs.Write.NamedExec(
		footprint.StmtInsertOrderClient,
		c)

	if err != nil {
		return err
	}

	return nil
}

// RetrieveOrder loads an order by its id.
func (env Env) RetrieveOrder(orderID string) (subs2.Order, error) {
	var order subs2.Order

	err := env.dbs.Read.Get(
		&order,
		subs2.StmtSelectOrder,
		orderID)

	if err != nil {
		return subs2.Order{}, err
	}

	// Set wx app id to the one used by native app pay if missing.
	if order.PaymentMethod == enum.PayMethodWx && order.WxAppID.IsZero() {
		order.WxAppID = null.StringFrom(wxAppNativeApp)
	}

	return order, nil
}

func (env Env) countOrders(ids ids.UserIDs) (int64, error) {
	var count int64
	err := env.dbs.Read.Get(
		&count,
		subs2.StmtCountOrders,
		ids.BuildFindInSet(),
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (env Env) listOrders(ids ids.UserIDs, p gorest.Pagination) ([]subs2.Order, error) {
	var orders = make([]subs2.Order, 0)
	err := env.dbs.Read.Select(
		&orders,
		subs2.StmtListOrders,
		ids.BuildFindInSet(),
		p.Limit,
		p.Offset(),
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (env Env) ListOrders(ids ids.UserIDs, p gorest.Pagination) (subs2.OrderList, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan subs2.OrderList)

	go func() {
		defer close(countCh)
		n, err := env.countOrders(ids)
		if err != nil {
			sugar.Error(err)
		}

		countCh <- n
	}()

	go func() {
		defer close(listCh)
		o, err := env.listOrders(ids, p)
		if err != nil {
			sugar.Error(err)
		}
		listCh <- subs2.OrderList{
			Total:      0,
			Pagination: gorest.Pagination{},
			Data:       o,
			Err:        err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return subs2.OrderList{}, listResult.Err
	}

	return subs2.OrderList{
		Total:      count,
		Pagination: p,
		Data:       listResult.Data,
	}, nil
}

func (env Env) orderHeader(orderID string) (subs2.Order, error) {
	var order subs2.Order

	err := env.dbs.Read.Get(
		&order,
		subs2.StmtOrderHeader,
		orderID)

	if err != nil {
		return subs2.Order{}, nil
	}

	return order, nil
}

func (env Env) orderTail(orderID string) (subs2.Order, error) {
	var order subs2.Order

	err := env.dbs.Read.Get(
		&order,
		subs2.StmtOrderTail,
		orderID)

	if err != nil {
		return subs2.Order{}, nil
	}

	return order, nil
}

type orderResult struct {
	order subs2.Order
	err   error
}

// LoadFullOrder retrieves an order by splitting a single row into two
// concurrent retrieval since for unknown reasons the DB does not
// respond if a row has two much columns.
func (env Env) LoadFullOrder(orderID string) (subs2.Order, error) {
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
		return subs2.Order{}, headerRes.err
	}

	if tailRes.err != nil {
		return subs2.Order{}, tailRes.err
	}

	return subs2.Order{
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
