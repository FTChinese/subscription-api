package account

// AuthResult contains the data retrieved from db after
// authenticating password for the specified email.
// If db returns ErrNoRows, it indicates the specified email does not
// exists.
// The the PasswordMatched field is false, it indicates the account
// exists but password is not correct.
type AuthResult struct {
	UserID          string `db:"user_id"`
	PasswordMatched bool   `db:"password_matched"`
}
