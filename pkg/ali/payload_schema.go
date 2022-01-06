package ali

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/smartwalle/alipay"
)

type RowKind string

const (
	RowKindWebhook    RowKind = "webhook"
	RowKindQueryOrder RowKind = "query_order"
)

type WebhookPayload struct {
	OrderID    string               `db:"order_id"`
	Payload    ColumnWebhookPayload `db:"payload"`
	CreatedUTC chrono.Time          `db:"created_utc"`
	RowKind    RowKind              `db:"kind"`
}

func NewWebhookPayload(n *alipay.TradeNotification) WebhookPayload {
	return WebhookPayload{
		OrderID: n.OutTradeNo,
		Payload: ColumnWebhookPayload{
			TradeNotification: n,
		},
		CreatedUTC: chrono.TimeNow(),
		RowKind:    RowKindWebhook,
	}
}

type OrderQueryPayload struct {
	OrderID    string           `db:"order_id"`
	Payload    ColumnOrderQuery `db:"payload"`
	CreatedUTC chrono.Time      `db:"created_utc"`
	RowKind    RowKind          `db:"kind"`
}

func NewOrderQueryPayload(p *alipay.AliPayTradeQueryResponse) OrderQueryPayload {
	return OrderQueryPayload{
		OrderID: p.AliPayTradeQuery.OutTradeNo,
		Payload: ColumnOrderQuery{
			AliPayTradeQueryResponse: p,
		},
		CreatedUTC: chrono.TimeNow(),
		RowKind:    RowKindQueryOrder,
	}
}

const StmtSavePayload = `
INSERT INTO premium.alipay_response
SET order_id = :order_id,
	payload = :payload,
	created_utc = :created_utc,
	kind = :kind
`
