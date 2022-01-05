package ali

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/smartwalle/alipay"
)

type ColumnOrderQuery struct {
	*alipay.AliPayTradeQueryResponse
}

func (p ColumnOrderQuery) Value() (driver.Value, error) {
	if p.AliPayTradeQueryResponse == nil {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (p *ColumnOrderQuery) Scan(src interface{}) error {
	if src == nil {
		*p = ColumnOrderQuery{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnOrderQuery
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
