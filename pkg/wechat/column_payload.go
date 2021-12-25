package wechat

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/objcoding/wxpay"
)

// ColumnPayload wraps wxpay.Params to save into JSON column.
// Not if you marshal the struct directly, the result
// if a map, but a struct with a field `Params`.
// This differs from embedding a struct.
type ColumnPayload struct {
	wxpay.Params
}

func (p ColumnPayload) Value() (driver.Value, error) {
	if p.Params == nil {
		return nil, nil
	}

	b, err := json.Marshal(p.Params)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (p *ColumnPayload) Scan(src interface{}) error {
	if src == nil {
		*p = ColumnPayload{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp map[string]string
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*p = ColumnPayload{
			Params: tmp,
		}
		return nil

	default:
		return errors.New("incompatible type to scan to ali.ColumnSDKParams")
	}
}
