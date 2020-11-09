package order_poller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type OrderPoller struct {
	db           *sqlx.DB
	logger       *zap.Logger
	aliPayClient subrepo.AliPayClient
	wxPayClients subrepo.WxPayClientStore
}

func NewOrderPoller(db *sqlx.DB, logger *zap.Logger) OrderPoller {

	aliApp := ali.MustInitApp()
	wxApps := wechat.MustGetPayApps()

	return OrderPoller{
		db:           db,
		logger:       logger,
		aliPayClient: subrepo.NewAliPayClient(aliApp, logger),
		wxPayClients: subrepo.NewWxClientStore(wxApps, logger),
	}
}

func (p OrderPoller) createOrderChannel() chan<- subs.Order {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()

	ch := make(chan subs.Order)

	go func() {
		defer close(ch)

		rows, err := p.db.Queryx(subs.StmtAliUnconfirmed)
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

		rows, err = p.db.Queryx(subs.StmtWxUnconfirmed)
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

func (p OrderPoller) verify(order subs.Order) {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()

	sugar.Infof("Start verifying order %s", order.ID)

	var payResult subs.PaymentResult
	var err error
	switch order.PaymentMethod {
	case enum.PayMethodWx:

	case enum.PayMethodAli:
		payResult, err = p.aliPayClient.VerifyPayment(order)
	}

	sugar.Infof("Payment result: %v", payResult)

	if err != nil {
		sugar.Error(err)
		return
	}

	if !payResult.IsOrderPaid() {
		sugar.Infof("Order %s is not paid", order.ID)
		return
	}

}
