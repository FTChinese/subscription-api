package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/sub"
)

func GetSubscription(subsID string) (*stripe.Subscription, error) {
	return sub.Get(subsID, nil)
}

// RefreshMembership refreshes an existing valid stripe membership.
func RefreshMembership(m reader.Membership, ss *stripe.Subscription) reader.Membership {
	periodEnd := dt.FromUnix(ss.CurrentPeriodEnd)

	m.ExpireDate = chrono.DateFrom(periodEnd.AddDate(0, 0, 1))
	m.Status, _ = enum.ParseSubsStatus(string(ss.Status))
	m.AutoRenewal = m.Status == enum.SubsStatusActive && IsAutoRenewal(ss)

	return m
}
