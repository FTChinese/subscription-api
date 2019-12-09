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

func (otx OrderTx) FreeUpgrade(builder *subscription.OrderBuilder) (subscription.Order, error) {

	if err := otx.PreviewUpgrade(builder); err != nil {
		return subscription.Order{}, err
	}

	if err := builder.Build(); err != nil {
		return subscription.Order{}, err
	}

	if !builder.CanBalanceCoverPlan() {
		return subscription.Order{}, subscription.ErrBalanceCannotCoverUpgrade
	}

	// Save order
	order, member, err := builder.FreeUpgradeOrder()
	if err != nil {
		return subscription.Order{}, err
	}

	// Save upgrade plan.
	upgradeSchema, _ := builder.UpgradeBalanceSchema()
	if err := otx.SaveUpgradeIntent(upgradeSchema); err != nil {
		return subscription.Order{}, err
	}

	// Save balance source.
	if err := otx.SaveProratedOrders(builder.ProratedOrdersSchema()); err != nil {
		return subscription.Order{}, err
	}

	if err := otx.UpdateMember(member); err != nil {
		return subscription.Order{}, err
	}

	return order, nil
}
