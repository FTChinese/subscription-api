package account

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/client"
	"github.com/guregu/null"
)

type FootprintSource string

const (
	FootprintSourceNull          FootprintSource = ""
	FootprintSourceLogin         FootprintSource = "login"
	FootprintSourceSignUp        FootprintSource = "signup"
	FootprintSourceVerification  FootprintSource = "email_verification"
	FootprintSourcePasswordReset FootprintSource = "password_reset"
)

type ClientFootprint struct {
	FtcID string `db:"ftc_id"`
	client.Client
	CreatedUTC  chrono.Time      `db:"created_utc"`
	Source      FootprintSource  `db:"source"`
	AuthMethod  enum.LoginMethod `db:"auth_method"` // Present wWhen Source is login.
	DeviceToken null.String      `db:"device_token"`
}

func NewFootprint(id string, client client.Client) ClientFootprint {
	return ClientFootprint{
		FtcID:      id,
		Client:     client,
		CreatedUTC: chrono.TimeNow(),
		Source:     "",
	}
}

// FromLogin set Source to login
func (c ClientFootprint) FromLogin() ClientFootprint {
	c.Source = FootprintSourceLogin
	return c
}

// FromSignUp set Source to signup
func (c ClientFootprint) FromSignUp() ClientFootprint {
	c.Source = FootprintSourceSignUp
	return c
}

// WithAuth adds the AuthMethod field and DeviceToken
// if logged in from mobile apps.
func (c ClientFootprint) WithAuth(method enum.LoginMethod, deviceToken null.String) ClientFootprint {
	c.AuthMethod = method
	c.DeviceToken = deviceToken
	return c
}

// FromVerification set the source to verification
func (c ClientFootprint) FromVerification() ClientFootprint {
	c.Source = FootprintSourceVerification
	return c
}

// FromPwReset sets the source to password reset request.
func (c ClientFootprint) FromPwReset() ClientFootprint {
	c.Source = FootprintSourcePasswordReset
	return c
}
