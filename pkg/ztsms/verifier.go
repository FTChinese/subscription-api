package ztsms

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"time"
)

// Verifier represents a row in the mobile_verifier
// table.
// Verification creation process:
// 1. Send mobile to API;
// 2. API user the mobile to create a verifier;
// 3. User should receive the verification code
// on device;
// 4. User input the code to client, and client
// submit the mobile and code together to API;
// 5. API verify the mobile and code.
type Verifier struct {
	Mobile     string      `db:"mobile_phone"`
	Code       string      `db:"sms_code"`
	ExpiresIn  int         `db:"expires_in"`
	CreatedUTC chrono.Time `db:"created_utc"`
	UsedUTC    chrono.Time `db:"used_utc"`
	FtcID      null.String `db:"ftc_id"`
}

// NewVerifier generates a new verification code for a mobile phone.
func NewVerifier(mobile string, ftcID null.String) Verifier {
	return Verifier{
		Mobile:     mobile,
		Code:       ids.SMSCode(),
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
