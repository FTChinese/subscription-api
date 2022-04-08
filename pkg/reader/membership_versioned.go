package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

const StmtVersionMembership = `
INSERT INTO premium.member_version
SET id = :snapshot_id,
	ante_change = :ante_change,
	created_by = :created_by,
	created_utc = :created_utc,
	b2b_transaction_id = :b2b_transaction_id,
	post_change = :post_change,
	retail_order_id = :retail_order_id
`

// MembershipVersioned stores a specific version of membership.
// Since membership is constantly changing, we keep all
// versions of modification in a dedicated table.
type MembershipVersioned struct {
	ID               string           `json:"id" db:"snapshot_id"`
	AnteChange       MembershipColumn `json:"anteChange" db:"ante_change"` // Membership before being changed
	CreatedBy        null.String      `json:"createdBy" db:"created_by"`
	CreatedUTC       chrono.Time      `json:"createdUtc" db:"created_utc"`
	B2BTransactionID null.String      `json:"b2bTransactionId" db:"b2b_transaction_id"`
	PostChange       MembershipColumn `json:"postChange" db:"post_change"`       // Membership after being changed.
	RetailOrderID    null.String      `json:"retailOderId" db:"retail_order_id"` // Only exists when user is performing renewal or upgrading.
}

func (v MembershipVersioned) WithCreator(name string) MembershipVersioned {
	v.CreatedBy = null.StringFrom(name)
	return v
}

// IsZero tests if a snapshot exists.
func (v MembershipVersioned) IsZero() bool {
	return v.ID == ""
}

func (v MembershipVersioned) WithB2BTxnID(id string) MembershipVersioned {
	v.B2BTransactionID = null.StringFrom(id)
	return v
}

// WithRetailOrderID sets the retail order id when taking a
// snapshot.
func (v MembershipVersioned) WithRetailOrderID(id string) MembershipVersioned {
	v.RetailOrderID = null.StringFrom(id)
	return v
}

// WithPriorVersion stores a previous version of membership
// before modification.
// It may not exist for newly created membership.
func (v MembershipVersioned) WithPriorVersion(m Membership) MembershipVersioned {
	if m.IsZero() {
		return v
	}

	v.AnteChange = MembershipColumn{m}

	return v
}

// Version takes a snapshots of the latest membership.
// Call WithPriorVersion method if you are updating an existing
// membership.
func (m Membership) Version(by Archiver) MembershipVersioned {
	if m.IsZero() {
		return MembershipVersioned{}
	}

	return MembershipVersioned{
		ID:               ids.SnapshotID(),
		AnteChange:       MembershipColumn{}, // Optional. Only exists if a previous version existed.
		CreatedBy:        null.StringFrom(by.String()),
		CreatedUTC:       chrono.TimeNow(),
		B2BTransactionID: null.String{},
		PostChange:       MembershipColumn{m},
		RetailOrderID:    null.String{},
	}
}

func (m Membership) Deleted(by Archiver) MembershipVersioned {
	return MembershipVersioned{
		ID:               ids.SnapshotID(),
		AnteChange:       MembershipColumn{m},
		CreatedBy:        null.StringFrom(by.String()),
		CreatedUTC:       chrono.TimeNow(),
		B2BTransactionID: null.String{},
		PostChange:       MembershipColumn{},
		RetailOrderID:    null.String{},
	}
}
