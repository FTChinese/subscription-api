package footprint

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

type Footprint struct {
	FtcID string `db:"ftc_id"`
	Client
	CreatedUTC  chrono.Time      `db:"created_utc"`
	Source      Source           `db:"source"`
	AuthMethod  enum.LoginMethod `db:"auth_method"` // Present wWhen Source is login.
	DeviceToken null.String      `db:"device_token"`
}

func New(id string, client Client) Footprint {
	return Footprint{
		FtcID:      id,
		Client:     client,
		CreatedUTC: chrono.TimeNow(),
		Source:     "",
	}
}

// FromLogin set Source to login
func (c Footprint) FromLogin() Footprint {
	c.Source = SourceLogin
	return c
}

// FromSignUp set Source to signup
func (c Footprint) FromSignUp() Footprint {
	c.Source = SourceSignUp
	return c
}

// WithAuth adds the AuthMethod field and DeviceToken
// if logged in from mobile apps.
func (c Footprint) WithAuth(method enum.LoginMethod, deviceToken null.String) Footprint {
	c.AuthMethod = method
	c.DeviceToken = deviceToken
	return c
}

// FromVerification set the source to verification
func (c Footprint) FromVerification() Footprint {
	c.Source = SourceVerification
	return c
}

// FromPwReset sets the source to password reset request.
func (c Footprint) FromPwReset() Footprint {
	c.Source = SourcePasswordReset
	return c
}
