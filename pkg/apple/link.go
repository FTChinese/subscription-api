package apple

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"strings"
)

// LinkInput defines the request body to link IAP to ftc account.
type LinkInput struct {
	FtcID        string `json:"ftcId" db:"ftc_user_id"`
	OriginalTxID string `json:"originalTxId" db:"original_transaction_id"`
	Force        bool   `json:"force"` // If FtcID is already linked to another apple id, it is not allowed to use this OriginalTxId's subscription unless Force is true.
}

func (i *LinkInput) Validate() *render.ValidationError {
	i.FtcID = strings.TrimSpace(i.FtcID)
	i.OriginalTxID = strings.TrimSpace(i.OriginalTxID)

	ve := validator.New("ftcId").Required().Validate(i.FtcID)
	if ve != nil {
		return ve
	}

	return validator.New("originalTxId").Required().Validate(i.OriginalTxID)
}

type LinkResult struct {
	Initial  bool                  // Is this initial link?
	Member   reader.Membership     // Updated membership or original membership if already linked and expiration date not changed.
	Snapshot reader.MemberSnapshot // Previous memberships. The original membership should be deleted, and then archived. If snapshot exists, them Touched must be always true.
}

type UnlinkResult struct {
	IAPSubs  Subscription
	Snapshot reader.MemberSnapshot
}

type LinkBuilder struct {
	Account    account.BaseAccount
	CurrentFtc reader.Membership
	CurrentIAP reader.Membership
	IAPSubs    Subscription
	Force      bool // If CurrentFtc is already linked to another Subscription, the link request should be denied unless Force is explicitly set to true.
}

// Build links IAP subscription to an existing FTC account.
// Only two cases are allowed to link:
//
// * Both sides refer to the same membership (including zero value);
// * IAP side is zero and FTC side non-zero but invalid.
//
// As long as link is allowed to proceed, two side cannot both have memberships simultaneously.
// We only need to take a snapshot of ftc side if it exists.
//
// | FTC\IAP     | None   | Not-Expired | Expired |
// | ----------- | ------ | ----------- | --------|
// | None        |  Y     |      N      |  N      |
// | Not-Expired |  N     |      N      |  N      |
// | Expired     |  Y     |      N      |  N      |
//
// Row 2 Column 2 has an exception:
// If payMethod is null, ftc side expire time is not after iap side, it is probably comes from IAP.
func (b LinkBuilder) Build() (LinkResult, error) {
	// Two sides get the same membership, or both zero values.
	if b.CurrentIAP.IsEqual(b.CurrentFtc) {
		// If one does not exist, both do not exist. We are creating a new membership in such case.
		if b.CurrentIAP.IsZero() {
			return LinkResult{
				Initial: true, // Send email
				Member: NewMembership(MembershipParams{
					UserID: b.Account.CompoundIDs(),
					Subs:   b.IAPSubs,
					AddOn:  addon.AddOn{},
				}),
				Snapshot: reader.MemberSnapshot{}, // Nothing to archive.
			}, nil
		}

		// Both sides not zero and they are equal, so this is
		// an existing iap membership.
		// If ftc side is older than IAP, update it.
		if b.CurrentFtc.ExpireDate.Before(b.IAPSubs.ExpiresDateUTC.Time) {
			return LinkResult{
				Initial: false,
				Member: NewMembership(MembershipParams{
					UserID: b.Account.CompoundIDs(),
					Subs:   b.IAPSubs,
					AddOn:  b.CurrentFtc.AddOn,
				}),
				Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
			}, nil
		}

		// This error indicates you do not need to handle it and send current membership back to client.
		return LinkResult{}, ErrAlreadyLinked
	}

	// If the two sides are not equal, they must be totally different memberships and the two sides cannot be both empty.
	//
	// a != b might indicates those cases:
	// 1. a == 0, b != 0;
	// 2. a != 0, b != 0;
	// 3. a != 0, b == 0.
	//
	// Case 1 and 2:
	// The presence of IAP side itself indicates it is already linked to an FTC account.
	// Now it is trying to link to another FTC account since the two sides are not the same one.
	// Such action should be denied regardless of whether the FTC side is valid or not, and it is mostly a fraudulent behavior: one apple id is trying to link to multiple Ftc Account.
	// If an exiting linked IAP is trying to switch the linked FTC account, it falls into this category and user should first perform unlink.
	if !b.CurrentIAP.IsZero() {
		return LinkResult{}, ErrIAPAlreadyLinked
	}

	// Case 3:
	// Current IAP side is empty, then FTC side must not be empty.
	// We need to consider whetherFTC side is created via another IAP.
	// Such case might arise when user is switching between 2 apple account
	// and both accounts have a subscription:
	// Apple Account A <--> Ftc Account A;
	// Apple Account B <-|-> Ftc Account A.
	// In such case we should deny it unless user is manually changing it and the `force` parameter should be provided.
	if b.CurrentFtc.IsIAP() {
		if b.Force {
			return LinkResult{
				Initial: true,
				Member: NewMembership(MembershipParams{
					UserID: b.Account.CompoundIDs(),
					Subs:   b.IAPSubs,
					AddOn:  b.CurrentFtc.AddOn,
				}),
				Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
			}, nil
		}

		return LinkResult{}, ErrFtcAlreadyLinked
	}

	// Now we can be sure that FTC side has membership and it does not come from IAP,
	// or `payMethod` field might be null due to data being changed manually.
	if !b.CurrentFtc.IsExpired() {
		// An edge case here: if the data is in legacy format and payMethod is null, which might be created by wxpay or
		// or alipay, or might be manually created by customer service, we could not determine whether the link should
		// be allowed or not.
		// In such case, we will compare the expiration date.
		// If apple's expiration date comes later, allow the FTC
		// side to be overridden; otherwise we shall keep the FTC
		// side intact.
		if b.isFtcLegacyFormat() {
			return LinkResult{
				Initial: true,
				Member: NewMembership(MembershipParams{
					UserID: b.Account.CompoundIDs(),
					Subs:   b.IAPSubs,
					AddOn:  b.CurrentFtc.AddOn,
				}),
				Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
			}, nil
		}

		// Carry-over invoice could be created here.
		return LinkResult{}, ErrFtcMemberValid
	}

	// FTC side is expired. If IAP subscription is also expired,
	// there's no need to touch data unless FTC side is in legacy format.
	if b.IAPSubs.IsExpired() {
		if b.isFtcLegacyFormat() {
			return LinkResult{
				Initial: true,
				Member: NewMembership(MembershipParams{
					UserID: b.Account.CompoundIDs(),
					Subs:   b.IAPSubs,
					AddOn:  b.CurrentFtc.AddOn,
				}),
				Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
			}, nil
		}

		return LinkResult{}, ErrIAPAlreadyExpired
	}

	// FTC side us expired while IAP subscription is not expired.
	// We can safely override the FTC side.
	return LinkResult{
		Initial: true,
		Member: NewMembership(MembershipParams{
			UserID: b.Account.CompoundIDs(),
			Subs:   b.IAPSubs,
			AddOn:  b.CurrentFtc.AddOn,
		}),
		Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
	}, nil
}

func (b LinkBuilder) isFtcLegacyFormat() bool {
	return b.CurrentFtc.PaymentMethod == enum.PayMethodNull && b.CurrentFtc.ExpireDate.AddDate(0, 0, -1).Before(b.IAPSubs.ExpiresDateUTC.Time)
}
