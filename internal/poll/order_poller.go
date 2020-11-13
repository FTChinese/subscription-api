package poll

import (
	"context"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/internal/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/poller"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const StmtAliUnconfirmed = subs.StmtOrderCols + `
FROM premium.log_ali_notification AS a
    LEFT JOIN premium.ftc_trade AS o
    ON a.ftc_order_id = o.trade_no
WHERE o.trade_no IS NOT NULL
	AND o.confirmed_utc IS NULL
	AND o.billing_cycle IS NOT NULL
    AND a.trade_status = 'TRADE_SUCCESS'`

const StmtWxUnconfirmed = subs.StmtOrderCols + `
FROM premium.log_wx_notification AS w
    LEFT JOIN premium.ftc_trade AS o
    ON w.ftc_order_id = o.trade_no
WHERE o.trade_no IS NOT NULL
	AND o.confirmed_utc IS NULL
	AND o.billing_cycle IS NOT NULL
    AND w.result_code = 'SUCCESS'`

type OrderPoller struct {
	db *sqlx.DB
	ftcpay.FtcPay
}

func NewOrderPoller(db *sqlx.DB, logger *zap.Logger) OrderPoller {
	return OrderPoller{
		db:     db,
		FtcPay: ftcpay.New(db, postoffice.New(config.MustGetHanqiConn()), logger),
	}
}

// retrieveOrders loads all orders created for alipay and wechat pay
// that are not confirmed but have valid webhook payloads.
// Returns a channel of order so that we could streaming each row
// rather than loading all rows in memory.
func (p OrderPoller) retrieveOrders() <-chan subs.Order {
	defer p.Logger.Sync()
	sugar := p.Logger.Sugar()

	ch := make(chan subs.Order)

	go func() {
		defer close(ch)

		rows, err := p.db.Queryx(StmtAliUnconfirmed)
		if err != nil {
			sugar.Error(err)
			return
		}

		order := subs.Order{}
		for rows.Next() {
			err := rows.StructScan(&order)
			if err != nil {
				sugar.Error(err)
				continue
			}

			sugar.Infof("%v\n", order)

			ch <- order
		}

		rows, err = p.db.Queryx(StmtWxUnconfirmed)
		if err != nil {
			sugar.Error(err)
			return
		}

		for rows.Next() {
			err := rows.StructScan(&order)
			if err != nil {
				sugar.Error(err)
				continue
			}

			sugar.Infof("%v\n", order)

			ch <- order
		}
	}()

	return ch
}

func (p OrderPoller) verify(order subs.Order) error {
	defer p.Logger.Sync()
	sugar := p.Logger.Sugar().With("orderId", order.ID)

	sugar.Info("Start verifying order")

	payResult, err := p.VerifyOrder(order)
	sugar.Infof("Payment result: %v", payResult)

	if err != nil {
		sugar.Error(err)
		return err
	}

	// Save/update payment result.
	err = p.SubsRepo.SavePayResult(payResult)
	if err != nil {
		sugar.Error(err)
	}

	if !payResult.IsOrderPaid() {
		sugar.Infof("Payment result is not success: %s", payResult.PaymentState)
		return err
	}

	_, cfmErr := p.ConfirmOrder(payResult, order)
	if cfmErr != nil {
		sugar.Error(cfmErr)
		return cfmErr
	}

	return nil
}

func (p OrderPoller) saveLog(l *poller.Log) error {
	_, err := p.db.NamedExec(poller.StmtSaveLog, l)
	if err != nil {
		return err
	}

	return nil
}

func (p OrderPoller) Start(dryRun bool) error {
	defer p.Logger.Sync()
	sugar := p.Logger.Sugar()
	ctx := context.Background()

	orderCh := p.retrieveOrders()

	pollerLog := poller.NewLog(poller.AppNameFtc)

	for order := range orderCh {
		if err := orderSem.Acquire(ctx, 1); err != nil {
			sugar.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		go func(o subs.Order) {
			pollerLog.IncTotal()

			defer orderSem.Release(1)

			if dryRun {
				return
			}

			err := p.verify(o)
			if err != nil {
				pollerLog.IncFailure()
			} else {
				pollerLog.IncSuccess()
			}

		}(order)
	}

	// Acquire all of the tokens to wait for any remaining workers to finish.
	//
	// If you are already waiting for the workers by some other means (such as an
	// errgroup.Group), you can omit this final Acquire call.
	if err := orderSem.Acquire(ctx, int64(maxWorkers)); err != nil {
		sugar.Infof("Failed to acquire semaphore: %v", err)
		return nil
	}

	pollerLog.EndUTC = chrono.TimeNow()

	err := savePollerLog(p.db, pollerLog)
	if err != nil {
		return err
	}

	sugar.Infof("Polling finished %v", pollerLog)
	return nil
}

func (p OrderPoller) Close() {
	defer p.Logger.Sync()
	sugar := p.Logger.Sugar()

	if err := p.db.Close(); err != nil {
		sugar.Error(err)
	}
}
