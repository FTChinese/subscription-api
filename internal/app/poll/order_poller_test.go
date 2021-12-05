package poll

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/poller"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestOrderPoller_createOrderChannel(t *testing.T) {

	p := NewOrderPoller(db.MockMySQL(), zaptest.NewLogger(t))

	orderCh := p.retrieveOrders()

	for order := range orderCh {
		t.Logf("%v", order)
	}
}

func mustGetOrder(db *sqlx.DB) subs.Order {
	var order subs.Order
	err := db.Get(&order, StmtAliUnconfirmed+` LIMIT 1`)
	if err != nil {
		panic(err)
	}

	return order
}

func TestOrderPoller_verify(t *testing.T) {
	poller := NewOrderPoller(db.MockMySQL(), zaptest.NewLogger(t))

	var order = mustGetOrder(poller.db)

	t.Logf("%v", order)

	err := poller.verify(order)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestOrderPoller_saveLog(t *testing.T) {
	pol := NewOrderPoller(db.MockMySQL(), zaptest.NewLogger(t))

	err := pol.saveLog(&poller.Log{
		Total:     10,
		Succeeded: 9,
		Failed:    1,
		StartUTC:  chrono.TimeNow(),
		EndUTC:    chrono.TimeNow(),
		AppName:   poller.AppNameFtc,
	})

	if err != nil {
		t.Error(err)
		return
	}
}

func TestOrderPoller_Start(t *testing.T) {
	p := NewOrderPoller(db.MockMySQL(), zaptest.NewLogger(t))

	err := p.Start(false)

	if err != nil {
		t.Error(err)
		return
	}

	p.Close()
}
