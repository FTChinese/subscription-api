package subs

import "errors"

var (
	// Reasons to deny upgrade.
	ErrNotUpgradeIntent = errors.New("upgrade could only be performed on an existing non-premium membership purchased via alipay or wxpay")
)
