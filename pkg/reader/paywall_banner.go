package reader

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"strings"
)

type BannerKind string

const (
	BannerKindDaily BannerKind = "daily"
	BannerKindPromo BannerKind = "promo"
)

type BannerJSON struct {
	ID              string      `json:"id"`
	Heading         string      `json:"heading"`
	SubHeading      null.String `json:"subHeading"`
	CoverURL        null.String `json:"coverUrl"`
	Content         null.String `json:"content"`
	Terms           null.String `json:"terms"`
	dt.ChronoPeriod             // Only exists for promo.
}

func (j *BannerJSON) Validate(k BannerKind) *render.ValidationError {
	j.Heading = strings.TrimSpace(j.Heading)
	j.CoverURL.String = strings.TrimSpace(j.CoverURL.String)
	j.SubHeading.String = strings.TrimSpace(j.SubHeading.String)
	j.Content.String = strings.TrimSpace(j.Content.String)
	j.Terms.String = strings.TrimSpace(j.Terms.String)

	if k == BannerKindPromo {
		if j.StartUTC.IsZero() {
			return &render.ValidationError{
				Message: "Start time is required",
				Field:   "startUtc",
				Code:    render.CodeMissing,
			}
		}

		if j.EndUTC.IsZero() {
			return &render.ValidationError{
				Message: "End time is requird",
				Field:   "endUtc",
				Code:    render.CodeMissing,
			}
		}

		if !j.EndUTC.After(j.StartUTC.Time) {
			return &render.ValidationError{
				Message: "End time should after start time",
				Field:   "endUtc",
				Code:    render.CodeInvalid,
			}
		}
	}

	return validator.New("heading").Required().Validate(j.Heading)
}

func (j BannerJSON) WithID(k BannerKind) BannerJSON {
	var id string

	switch k {
	case BannerKindDaily:
		id = ids.BannerID()

	case BannerKindPromo:
		id = ids.PromoID()
	}

	j.ID = id

	return j
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (j BannerJSON) Value() (driver.Value, error) {
	if j.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (j *BannerJSON) Scan(src interface{}) error {
	if src == nil {
		*j = BannerJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp BannerJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*j = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to BannerJSON")
	}
}
