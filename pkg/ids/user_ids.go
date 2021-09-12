package ids

import (
	"errors"
	"github.com/guregu/null"
	"strings"
)

// UserIDs is used to identify an FTC user.
// A user might have an ftc uuid, or a wechat union id,
// or both.
// This type structure is used to ensure unique constraint
// for SQL columns that cannot be both null since SQL do not
// have a mechanism to do UNIQUE INDEX on two columns while
// keeping either of them nullable.
// A user's compound id is taken from either ftc uuid or
// wechat id, with ftc id taking precedence.
type UserIDs struct {
	CompoundID string      `json:"-" db:"compound_id"`
	FtcID      null.String `json:"ftcId" db:"ftc_id"`
	UnionID    null.String `json:"unionId" db:"union_id"`
}

func NewFtcUserID(id string) UserIDs {
	return UserIDs{
		CompoundID: id,
		FtcID:      null.StringFrom(id),
		UnionID:    null.String{},
	}
}

func (u UserIDs) Normalize() (UserIDs, error) {
	if u.FtcID.IsZero() && u.UnionID.IsZero() {
		return u, errors.New("ftcID and unionID should not both be null")
	}

	if u.FtcID.Valid {
		u.CompoundID = u.FtcID.String
		return u, nil
	}

	u.CompoundID = u.UnionID.String
	return u, nil
}

func (u UserIDs) MustNormalize() UserIDs {
	ids, err := u.Normalize()
	if err != nil {
		panic(err)
	}

	return ids
}

// BuildFindInSet builds a value to be used in MySQL
// function FIND_IN_SET(str, strlist) so that find
// a user's data by both ftc id and union id.
func (u UserIDs) BuildFindInSet() string {
	strList := make([]string, 0)

	if u.FtcID.Valid {
		strList = append(strList, u.FtcID.String)
	}

	if u.UnionID.Valid {
		strList = append(strList, u.UnionID.String)
	}

	return strings.Join(strList, ",")
}

func (u UserIDs) IDSlice() []string {
	strList := make([]string, 0)

	if u.FtcID.Valid {
		strList = append(strList, u.FtcID.String)
	}

	if u.UnionID.Valid {
		strList = append(strList, u.UnionID.String)
	}

	return strList
}
