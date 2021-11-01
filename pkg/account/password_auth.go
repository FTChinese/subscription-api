package account

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
)

// AuthResult contains the data retrieved from db after
// authenticating password for the specified email or id.
// If db returns ErrNoRows, it indicates the specified email does not
// exist.
// If the PasswordMatched field is false, it indicates the account
// exists but password is not correct.
type AuthResult struct {
	UserID          string `db:"user_id"`
	PasswordMatched bool   `db:"password_matched"`
}

// IDCredentials holds a user's password.
// It could be used either to hold current password,
// or a new password to update.
type IDCredentials struct {
	FtcID    string `json:"-" db:"ftc_id"`
	Password string `json:"password" db:"password"`
}

func (c IDCredentials) Validate() *render.ValidationError {
	return validator.EnsurePassword(c.Password)
}
