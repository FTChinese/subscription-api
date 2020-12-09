package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type APIError struct {
	FtcID        string      `db:"ftc_user_id"`
	ChargeID     null.String `db:"charge_id"`
	ErrorCode    null.String `db:"error_code"`
	HttpStatus   null.Int    `db:"http_status"`
	ErrorMessage string      `db:"error_message"`
	Parameter    null.String `db:"parameter"`
	RequestID    null.String `db:"request_id"`
	ErrorType    string      `db:"error_type"`
	CreatedUTC   chrono.Time `db:"created_utc"`
}

func NewAPIError(ftcID string, e *stripe.Error) APIError {
	return APIError{
		FtcID:        ftcID,
		ChargeID:     null.NewString(e.ChargeID, e.ChargeID != ""),
		ErrorCode:    null.NewString(string(e.Code), e.Code != ""),
		HttpStatus:   null.NewInt(int64(e.HTTPStatusCode), e.HTTPStatusCode != 0),
		ErrorMessage: e.Msg,
		Parameter:    null.NewString(e.Param, e.Param != ""),
		RequestID:    null.NewString(e.RequestID, e.RequestID != ""),
		ErrorType:    string(e.Type),
		CreatedUTC:   chrono.Time{},
	}
}
