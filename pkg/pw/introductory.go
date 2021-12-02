package pw

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/guregu/null"
)

type Introductory struct {
	StripePriceID null.String `json:"stripePriceId"`
}

type IntroductoryJSON struct {
	Introductory
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (j IntroductoryJSON) Value() (driver.Value, error) {

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (j *IntroductoryJSON) Scan(src interface{}) error {
	if src == nil {
		*j = IntroductoryJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp IntroductoryJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*j = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to IntroductoryJSON")
	}
}
