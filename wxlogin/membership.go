package wxlogin

import (
	"time"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Membership contains user's subscription data.
// If membership is not bound to another account yet,
// UserID should not be empty if membership is purcahsed with FTC account while UnionID must be null;
// UserID should not be empty and UnionID should not be null, and their values are equal if membership is purchased with Wechat account;
// UserID should not be empty and UnionID should not be null, and their values should be different if FTC account is already bound to Wechat account.
// If membership is not bound to another account yet, either UserID or UnionID should be empty, if membership actually present.
type Membership struct {
	UserID     string      `json:"-"`
	UnionID    null.String `json:"-"`
	Tier       enum.Tier   `json:"tier"`
	Cycle      enum.Cycle  `json:"billingCycle"`
	ExpireDate util.Date   `json:"expireDate"`
}

// IsExpired test if a membership is expired.
func (m Membership) IsExpired() bool {
	return m.ExpireDate.Before(time.Now())
}

// IsEqualTo tests if two memberships are the same one.
// If return false, it indicates the two accounts are bound to 3rd accounts, or not bound to any account.
func (m Membership) IsEqualTo(other Membership) bool {
	return m.UserID == other.UserID
}

// IsCoupled tests if a membership is bound to another account.
// Any membership that is coupled to another one should deny merge request.
func (m Membership) IsCoupled() bool {
	return m.UnionID.Valid && (m.UnionID.String != m.UserID)
}

// IsFromFTC tests if this membership is purchased by ftc account
func (m Membership) IsFromFTC() bool {
	return !m.UnionID.Valid
}

// IsFromWx tests if this membership is purchased by wechat account
func (m Membership) IsFromWx() bool {
	return m.UnionID.Valid && (m.UnionID.String == m.UserID)
}

// IsEmpty tests if a membership is empty.
// Empty membership is defined as UserID == "" since the ftc_vip.vip_id column must have a value.
func (m Membership) IsEmpty() bool {
	return m.UserID == ""
}

// Merge merges two membership into one
// This is used if both FTC account and Wechat account of a user
// owns membership.
// One of the membership should be already expired or both expired.
// If both memberships are still valid, deny the request.
func (m Membership) Merge(other Membership) Membership {
	var merged Membership

	if m.ExpireDate.After(other.ExpireDate.Time) {
		merged = m
	} else {
		merged = other
	}

	if m.IsFromFTC() {
		merged.UserID = m.UserID
	} else if other.IsFromFTC() {
		merged.UserID = m.UserID
	}

	if m.IsFromWx() {
		merged.UnionID = m.UnionID
	} else if other.IsFromWx() {
		merged.UnionID = other.UnionID
	}

	return merged
}

// Pick picks from two membership the one that is not empty.
// The returned membership is missing either ftc user id or
// wechat union id.
func (m Membership) Pick(other Membership) Membership {
	if m.IsEmpty() && other.IsEmpty() {
		return m
	}

	if !m.IsEmpty() && !other.IsEmpty() {
		return m.Merge(other)
	}

	if !m.IsEmpty() {
		return m
	}

	return other
}
