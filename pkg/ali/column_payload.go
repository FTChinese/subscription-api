package ali

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/smartwalle/alipay"
)

type ColumnWebhookPayload struct {
	*alipay.TradeNotification
}

func (p ColumnWebhookPayload) Value() (driver.Value, error) {
	if p.TradeNotification == nil || p.TradeNo == "" {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (p *ColumnWebhookPayload) Scan(src interface{}) error {
	if src == nil {
		*p = ColumnWebhookPayload{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnWebhookPayload
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to ali.ColumnSDKParams")
	}
}
