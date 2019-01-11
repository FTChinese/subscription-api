package model

import "errors"

// Errors for controller to identify actions to take.
var (
	ErrOrderNotFound    = errors.New("Subscripiton order is not found")
	ErrAlreadyConfirmed = errors.New("Subscription order is already confirmed")
	ErrPriceMismatch    = errors.New("Price does not match")
)
