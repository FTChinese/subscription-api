package ztsms

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"time"
)

type VerifierParams struct {
	Mobile      string `json:"mobile"`
	Code        string `json:"code"`
	DeviceToken string `json:"deviceToken"`
}

func (p VerifierParams) ValidateMobile() *render.ValidationError {
	ok := validator.IsMobile(p.Mobile)
	if !ok {
		return &render.ValidationError{
			Message: "Invalid mobile number",
			Field:   "code",
			Code:    render.CodeInvalid,
		}
	}

	return nil
}

func (p VerifierParams) Validate() *render.ValidationError {
	ve := p.ValidateMobile()
	if ve != nil {
		return ve
	}

	return validator.New("code").
		MinLen(6).
		MaxLen(6).
		Required().
		Validate(p.Code)
}

type Verifier struct {
	Mobile     string      `db:"mobile_phone"`
	Code       string      `db:"sms_code"`
	ExpiresIn  int         `db:"expires_in"`
	CreatedUTC chrono.Time `db:"created_utc"`
	UsedUTC    chrono.Time `db:"used_utc"`
	FtcID      null.String `db:"ftc_id"`
}

// NewPhoneVerifier generates a new verification code for a mobile phone.
func NewVerifier(mobile string, ftcID null.String) Verifier {
	return Verifier{
		Mobile:     mobile,
		Code:       pkg.SMSCode(),
		ExpiresIn:  5 * 60,
		CreatedUTC: chrono.TimeNow(),
		UsedUTC:    chrono.Time{},
		FtcID:      ftcID,
	}
}

// WithUsed sets when a code is used so that it shouldn't be used again.
func (v Verifier) WithUsed() Verifier {
	v.UsedUTC = chrono.TimeNow()
	return v
}

func (v Verifier) Valid() bool {
	return v.CreatedUTC.Add(time.Duration(v.ExpiresIn) * time.Second).
		After(time.Now())
}
