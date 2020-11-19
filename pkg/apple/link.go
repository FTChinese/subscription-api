package apple

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/validator"
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
	Notify   bool                  // Whether we should tell user the link. If they are already linked, don't notify.
	Touched  bool                  // If the membership is actually modified. If link already done, this should be false and  Member should not be inserted to db.
	Member   reader.Membership     // Updated membership or original membership if already linked and expiration date not changed.
	Snapshot reader.MemberSnapshot // Previous memberships. The original membership should be deleted, and then archived. If snapshot exists, them Touched must be always true.
}

type UnlinkResult struct {
	IAPSubs  Subscription
	Snapshot reader.MemberSnapshot
}

type LinkBuilder struct {
	Account    reader.FtcAccount
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
				Notify:   true, // Send email
				Touched:  true, // Membership changed.
				Member:   b.IAPSubs.NewMembership(b.Account.MemberID()),
				Snapshot: reader.MemberSnapshot{}, // Nothing to archive.
			}, nil
		}

		// Current membership exists.
		// If expiration date not changed, tell caller do nothing.
		if b.CurrentIAP.ExpireDate.Equal(b.IAPSubs.ExpiresDateUTC.Time) {
			return LinkResult{
				Notify:   false,                   // Linked in prior session. Do not resend email
				Touched:  false,                   // Membership not modified.
				Member:   b.CurrentIAP,            // Will be sent to client.
				Snapshot: reader.MemberSnapshot{}, // Nothing changed. No archiving.
			}, nil
		}

		// If expiration date changed, updated it based on subscription.
		return LinkResult{
			Notify:   false,
			Touched:  true,
			Member:   b.IAPSubs.NewMembership(b.Account.MemberID()),
			Snapshot: b.CurrentIAP.Snapshot(reader.ArchiverAppleLink), // Archive current membership.
		}, nil
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
	// We need to consider 2 cases here:
	// Case 3-1:
	// FTC side is created via another IAP. Such case might arise when user is switching between 2 apple account
	// and both accounts have a subscription:
	// Apple Account A <--> Ftc Account A;
	// Apple Account B <-|-> Ftc Account A.
	// In such case we should deny it and user should manually unlink that IAP before linking to this one.
	if b.CurrentFtc.IsIAP() {
		if b.Force {
			return LinkResult{
				Notify:   true,
				Touched:  true,
				Member:   b.IAPSubs.NewMembership(b.Account.MemberID()),
				Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
			}, nil
		}

		return LinkResult{}, ErrFtcAlreadyLinked
	}

	// Case 3-2:
	// FTC side is non-IAP.
	// Then check whether it is expired.
	// If the FTC side is still valid, merging is not allowed since it will override valid data.

	if !b.CurrentFtc.IsExpired() {
		// An edge case here: if the data is in legacy format and payMethod is null, which might be created by wxpay or
		// or alipay, or might be manually created by customer service, we could not determine whether the linked should
		// be allowed or not.
		// In such case, we will compare the expiration date.
		// If apple's expiration date comes later, allow the FTC
		// side to be overridden; otherwise we shall keep the FTC
		// side intact.
		if b.CurrentFtc.PaymentMethod == enum.PayMethodNull {
			if b.CurrentFtc.ExpireDate.Before(b.IAPSubs.ExpiresDateUTC.Time) {
				return LinkResult{
					Notify:   true,
					Touched:  true,
					Member:   b.IAPSubs.NewMembership(b.Account.MemberID()),
					Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
				}, nil
			}
		}

		return LinkResult{}, ErrFtcMemberValid
	}

	return LinkResult{
		Notify:   true,
		Touched:  true,
		Member:   b.IAPSubs.NewMembership(b.Account.MemberID()),
		Snapshot: b.CurrentFtc.Snapshot(reader.ArchiverAppleLink),
	}, nil
}

func ConvertLinkErr(err error) (*render.ValidationError, bool) {
	switch err {
	// Multiple FTC accounts linking to single IAP.
	case ErrIAPAlreadyLinked:
		return &render.ValidationError{
			Message: "Apple subscription is already claimed by another ftc account.",
			Field:   "originalTxId",
			Code:    "linked_to_other_ftc",
		}, true

	// Single FTC account linking to multiple IAP
	case ErrFtcAlreadyLinked:
		return &render.ValidationError{
			Message: "FTC account is already linked to another Apple subscription",
			Field:   "ftcId",
			Code:    "linked_to_other_iap",
		}, true

	case ErrFtcMemberValid:
		return &render.ValidationError{
			Message: "FTC account already has a valid membership via non-Apple channel",
			Field:   "ftcId",
			Code:    "has_valid_non_iap",
		}, true

	default:
		return nil, false
	}
}
