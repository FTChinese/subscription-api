package reader

import (
	"errors"
	"github.com/guregu/null"
	"strings"
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
	CompoundID string      `json:"-" db:"compound_id"`
	FtcID      null.String `json:"ftcId" db:"ftc_id"`
	UnionID    null.String `json:"unionId" db:"union_id"`
}

func NewFtcUserID(id string) MemberID {
	return MemberID{
		CompoundID: id,
		FtcID:      null.StringFrom(id),
		UnionID:    null.String{},
	}
}

func (m MemberID) Normalize() (MemberID, error) {
	if m.FtcID.IsZero() && m.UnionID.IsZero() {
		return m, errors.New("ftcID and unionID should not both be null")
	}

	if m.FtcID.Valid {
		m.CompoundID = m.FtcID.String
		return m, nil
	}

	m.CompoundID = m.UnionID.String
	return m, nil
}

func (m MemberID) MustNormalize() MemberID {
	ids, err := m.Normalize()
	if err != nil {
		panic(err)
	}

	return ids
}

// BuildFindInSet builds a value to be used in MySQL
// function FIND_IN_SET(str, strlist) so that find
// a user's data by both ftc id and union id.
func (m MemberID) BuildFindInSet() string {
	strList := make([]string, 0)

	if m.FtcID.Valid {
		strList = append(strList, m.FtcID.String)
	}

	if m.UnionID.Valid {
		strList = append(strList, m.UnionID.String)
	}

	return strings.Join(strList, ",")
}

func (m MemberID) IDSlice() []string {
	strList := make([]string, 0)

	if m.FtcID.Valid {
		strList = append(strList, m.FtcID.String)
	}

	if m.UnionID.Valid {
		strList = append(strList, m.UnionID.String)
	}

	return strList
}
