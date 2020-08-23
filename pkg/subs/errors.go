package subs

import "errors"

var (
	// In case the plan is not set in OrderBuilder
	ErrInvalidPlan = errors.New("no pricing plan is provided")

	// Reasons to deny upgrade.
	ErrUpgradeInvalid            = errors.New("upgrade could only be performed on an existing non-premium membership purchased via alipay or wxpay")
	ErrBalanceCannotCoverUpgrade = errors.New("current balance is not enough to cover upgrading cost")
)
