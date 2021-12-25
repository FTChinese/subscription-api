package wxpay

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/go-pay/gopay/wechat"
)

type RowKind string

const (
	RowKindOrderResponse RowKind = "order_response"
)

const StmtSaveResponse = `
INSERT INTO premium.wxpay_response
SET order_id = :order_id,
	payload = :payload,
	created_utc = :created_utc,
	kind = :kind
`

type ColumnOrderResponse struct {
	wechat.UnifiedOrderResponse
}

func (c ColumnOrderResponse) Value() (driver.Value, error) {
	b, err := json.Marshal(c.UnifiedOrderResponse)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (c *ColumnOrderResponse) Scan(src interface{}) error {
	if src == nil {
		*c = ColumnOrderResponse{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp wechat.UnifiedOrderResponse
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*c = ColumnOrderResponse{
			UnifiedOrderResponse: tmp,
		}
		return nil

	default:
		return errors.New("incompatible type to scan to ColumnOrderResponse")
	}
}

type OrderResponseSchema struct {
	OrderID    string              `db:"order_id"`
	Payload    ColumnOrderResponse `db:"payload"`
	CreatedUTC chrono.Time         `db:"created_utc"`
	RowKind    RowKind             `db:"kind"`
}

func NewOrderResponseSchema(id string, payload wechat.UnifiedOrderResponse) OrderResponseSchema {
	return OrderResponseSchema{
		OrderID: id,
		Payload: ColumnOrderResponse{
			UnifiedOrderResponse: payload,
		},
		CreatedUTC: chrono.TimeNow(),
		RowKind:    RowKindOrderResponse,
	}
}
