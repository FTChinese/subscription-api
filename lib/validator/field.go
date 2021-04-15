package validator

import "github.com/FTChinese/go-rest/render"

func EnsureEmail(email string) *render.ValidationError {
	return New("email").
		Required().
		MaxLen(64).
		Email().
		Validate(email)
}

func EnsurePassword(pw string) *render.ValidationError {
	return New("password").
		Required().
		MaxLen(64).
		MinLen(8).
		Validate(pw)
}
