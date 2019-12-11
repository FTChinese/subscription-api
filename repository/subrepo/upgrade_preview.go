package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"time"
)

// See errors returned from Membership.PermitAliWxUpgrade.
func (otx OrderTx) PreviewUpgrade(builder *subscription.OrderBuilder) error {

	member, err := otx.RetrieveMember(builder.GetReaderID())
	if err != nil {
		return err
	}

	if !member.PermitAliWxUpgrade() {
		return subscription.ErrUpgradeInvalid
	}

	orders, err := otx.FindBalanceSources(builder.GetReaderID())
	if err != nil {
		return err
	}

	wallet := subscription.NewWallet(orders, time.Now())

	builder.SetMembership(member).
		SetWallet(wallet)

	return nil
}
