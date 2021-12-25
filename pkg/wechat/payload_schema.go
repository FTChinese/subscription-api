package wechat

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/objcoding/wxpay"
)

type RowKind string

const (
	RowKindCreateOrder RowKind = "create_order"
	RowKindWebhook     RowKind = "webhook"
	RowKindQueryOrder  RowKind = "query_order"
)

const StmtSavePayload = `
INSERT INTO premium.wxpay_response
SET order_id = :order_id,
	payload = :payload,
	created_utc = :created_utc,
	kind = :kind
`

type PayloadSchema struct {
	OrderID    string        `db:"order_id"`
	Payload    ColumnPayload `db:"payload"`
	CreatedUTC chrono.Time   `db:"created_utc"`
	RowKind    RowKind       `db:"kind"`
}

func (s PayloadSchema) WithKind(k RowKind) PayloadSchema {
	s.RowKind = k

	return s
}

func NewPayloadSchema(id string, body wxpay.Params) PayloadSchema {
	return PayloadSchema{
		OrderID: id,
		Payload: ColumnPayload{
			Params: body,
		},
		CreatedUTC: chrono.TimeNow(),
		RowKind:    "",
	}
}
