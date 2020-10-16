package reader

import (
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go"
)

const stmtColsSetSubsLink = `
stripe_subscription_id = IFNULL(:stripe_subs_id, stripe_subscription_id),
apple_original_tx_id = IFNULL(:apple_original_tx_id, apple_original_tx_id),
b2b_licence_id = IFNULL(:b2b_licence_id, b2b_licence_id)
`

const StmtSaveSubsLink = `
INSERT INTO premium.subs_link
SET ftc_user_id = :ftc_user_id,
` + stmtColsSetSubsLink + `
ON DUPLICATE KEY UPDATE
` + stmtColsSetSubsLink

// SubsLink links ftc uuid to ids from various subscription channel.
type SubsLink struct {
	FtcID             string      `db:"ftc_user_id"`
	StripeSubsID      null.String `db:"stripe_subs_id"`
	AppleOriginalTxID null.String `json:"apple_original_tx_id"`
	B2BLicenceID      null.String `json:"b2b_licence_id"`
}

func NewSubsLinkStripe(ftcID string, ss *stripeSdk.Subscription) SubsLink {
	return SubsLink{
		FtcID:             ftcID,
		StripeSubsID:      null.StringFrom(ss.ID),
		AppleOriginalTxID: null.String{},
		B2BLicenceID:      null.String{},
	}
}

func NewSubsLink(m Membership) SubsLink {
	return SubsLink{
		FtcID:             m.FtcID.String,
		StripeSubsID:      m.StripeSubsID,
		AppleOriginalTxID: m.AppleSubsID,
		B2BLicenceID:      m.B2BLicenceID,
	}
}
