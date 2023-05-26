package subrepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// CreateOrder creates an order and save it to db.
// This version is applicable to all users, regardless of how their current membership is purchased. It no longer calculates user's
// current account balance.
// For upgrading to premium with valid standard subscription,
// the remaining days is converted to add-on.
// For Stripe and IAP, the purchase is taken as add-on directly.
func (env Env) CreateOrder(cart reader.ShoppingCart) (ftcpay.PaymentIntent, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	otx, err := env.BeginOrderTx()
	if err != nil {
		sugar.Error(err)
		return ftcpay.PaymentIntent{}, err
	}

	// Step 1: Retrieve membership for this user.
	// The membership might be empty but the value is
	// valid.
	sugar.Infof("Start retrieving membership for reader %+v", cart.Account.CompoundIDs())
	member, err := otx.RetrieveMember(cart.Account.CompoundID())
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return ftcpay.PaymentIntent{}, err
	}
	sugar.Infof("Membership retrieved %+v", member)

	cart, err = cart.WithMember(member)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return ftcpay.PaymentIntent{}, err
	}

	// First calculate a Checkout instance.
	// Then see if the Offer field has Recurring.
	// If not recurring, retrieve from db the usage history
	// by user id and offer id.
	// If not found, continue to calculate payment intent;
	// if found,
	pi, err := ftcpay.NewPaymentIntent(cart)
	if err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return ftcpay.PaymentIntent{}, err
	}

	// Step 4: Save this order.
	if err := otx.SaveOrder(pi.Order); err != nil {
		sugar.Error(err)
		_ = otx.Rollback()
		return ftcpay.PaymentIntent{}, err
	}
	sugar.Infof("Order saved %s", pi.Order.ID)

	if err := otx.Commit(); err != nil {
		sugar.Error(err)
		return ftcpay.PaymentIntent{}, err
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
func (env Env) RetrieveOrder(orderID string) (ftcpay.Order, error) {
	var order ftcpay.Order

	err := env.dbs.Read.Get(
		&order,
		ftcpay.StmtSelectOrder,
		orderID)

	if err != nil {
		return ftcpay.Order{}, err
	}

	return order, nil
}

func (env Env) countOrders(uids ids.UserIDs) (int64, error) {
	var count int64
	err := env.dbs.Read.Get(
		&count,
		ftcpay.StmtCountOrders,
		uids.BuildFindInSet(),
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (env Env) listOrders(ids ids.UserIDs, p gorest.Pagination) ([]ftcpay.Order, error) {
	var orders = make([]ftcpay.Order, 0)
	err := env.dbs.Read.Select(
		&orders,
		ftcpay.StmtListOrders,
		ids.BuildFindInSet(),
		p.Limit,
		p.Offset(),
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (env Env) ListOrders(ids ids.UserIDs, p gorest.Pagination) (pkg.PagedData[ftcpay.Order], error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	countCh := make(chan int64)
	listCh := make(chan pkg.AsyncResult[[]ftcpay.Order])

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
		listCh <- pkg.AsyncResult[[]ftcpay.Order]{
			Value: o,
			Err:   err,
		}
	}()

	count, listResult := <-countCh, <-listCh

	if listResult.Err != nil {
		return pkg.PagedData[ftcpay.Order]{}, listResult.Err
	}

	return pkg.PagedData[ftcpay.Order]{
		Total:      count,
		Pagination: p,
		Data:       listResult.Value,
	}, nil
}

func (env Env) orderHeader(orderID string) (ftcpay.Order, error) {
	var order ftcpay.Order

	err := env.dbs.Read.Get(
		&order,
		ftcpay.StmtOrderHeader,
		orderID)

	if err != nil {
		return ftcpay.Order{}, nil
	}

	return order, nil
}

func (env Env) orderTail(orderID string) (ftcpay.Order, error) {
	var order ftcpay.Order

	err := env.dbs.Read.Get(
		&order,
		ftcpay.StmtOrderTail,
		orderID)

	if err != nil {
		return ftcpay.Order{}, nil
	}

	return order, nil
}

type orderResult struct {
	value ftcpay.Order
	err   error
}

// LoadFullOrder retrieves an order by splitting a single row into two
// concurrent retrieval since for unknown reasons the DB does not
// respond if a row has two much columns.
func (env Env) LoadFullOrder(orderID string) (ftcpay.Order, error) {
	headerCh := make(chan orderResult)
	tailCh := make(chan orderResult)

	go func() {
		defer close(headerCh)

		order, err := env.orderHeader(orderID)
		headerCh <- orderResult{
			value: order,
			err:   err,
		}
	}()

	go func() {
		defer close(tailCh)

		order, err := env.orderTail(orderID)
		tailCh <- orderResult{
			value: order,
			err:   err,
		}
	}()

	headerRes, tailRes := <-headerCh, <-tailCh
	if headerRes.err != nil {
		return ftcpay.Order{}, headerRes.err
	}

	if tailRes.err != nil {
		return ftcpay.Order{}, tailRes.err
	}

	return headerRes.value.MergeTail(tailRes.value), nil
}
