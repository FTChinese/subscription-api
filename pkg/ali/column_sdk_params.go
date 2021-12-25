package ali

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ColumnSDKParams struct {
	SDKParams
}

func (p ColumnSDKParams) Value() (driver.Value, error) {
	if p.IsZero() {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (p *ColumnSDKParams) Scan(src interface{}) error {
	if src == nil {
		*p = ColumnSDKParams{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ColumnSDKParams
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
