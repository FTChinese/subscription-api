package reader

import (
	"errors"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/query"
)

// FtcID is used to identify an FTC user.
// A user might have an ftc uuid, or a wechat union id,
// or both.
// This type structure is used to ensure unique constraint
// for SQL columns that cannot be both null since SQL do not
// have a mechanism to do UNIQUE INDEX on two columns while
// keeping either of them nullable.
// A user's compound id is taken from either ftc uuid or
// wechat id, with ftc id taking precedence.
type MemberID struct {
	CompoundID string      `json:"-" db:"sub_compound_id"`
	FtcID      null.String `json:"-" db:"sub_ftc_id"`
	UnionID    null.String `json:"-" db:"sub_union_id"`
}

func NewMemberID(ftcID, unionID string) (MemberID, error) {
	id := MemberID{
		FtcID:   null.NewString(ftcID, ftcID != ""),
		UnionID: null.NewString(unionID, unionID != ""),
	}

	if ftcID != "" {
		id.CompoundID = ftcID
	} else if unionID != "" {
		id.CompoundID = unionID
	} else {
		return id, errors.New("ftcID and unionID should not both be null")
	}
	return id, nil
}

// MemberColumn determines which column will be used to
// retrieve membership.
func (i MemberID) MemberColumn() query.MemberCol {
	if i.FtcID.Valid {
		return query.MemberColCompoundID
	}

	if i.UnionID.Valid {
		return query.MemberColUnionID
	}

	return query.MemberColCompoundID
}
