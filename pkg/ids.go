package pkg

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/rand"
	"github.com/guregu/null"
	"strings"
)

// GenerateOrderID creates an order memberID.
// The memberID has a total length of 18 chars.
// If we use this generator:
// `FT` takes 2, followed by year-month-date-hour-minute
// FT201905191139, then only 4 chars left for random number
// 2^16 = 65536, which means only 60000 order could be created every minute.
// To leave enough space for random number, 8 chars might be reasonable - 22 chars totally.
// If we use current random number generator:
// 2 ^ 64 = 1.8 * 10^19 orders.
func OrderID() (string, error) {

	id, err := gorest.RandomHex(8)
	if err != nil {
		return "", err
	}

	return "FT" + strings.ToUpper(id), nil
}

func MustOrderID() string {
	id, err := OrderID()
	if err != nil {
		panic(err)
	}

	return id
}

func SnapshotID() string {
	return "snp_" + rand.String(12)
}

func InvoiceID() string {
	return "inv_" + rand.String(12)
}

// FtcID is used to identify an FTC user.
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

func (m UserIDs) Normalize() (UserIDs, error) {
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

func (m UserIDs) MustNormalize() UserIDs {
	ids, err := m.Normalize()
	if err != nil {
		panic(err)
	}

	return ids
}

// BuildFindInSet builds a value to be used in MySQL
// function FIND_IN_SET(str, strlist) so that find
// a user's data by both ftc id and union id.
func (m UserIDs) BuildFindInSet() string {
	strList := make([]string, 0)

	if m.FtcID.Valid {
		strList = append(strList, m.FtcID.String)
	}

	if m.UnionID.Valid {
		strList = append(strList, m.UnionID.String)
	}

	return strings.Join(strList, ",")
}

func (m UserIDs) IDSlice() []string {
	strList := make([]string, 0)

	if m.FtcID.Valid {
		strList = append(strList, m.FtcID.String)
	}

	if m.UnionID.Valid {
		strList = append(strList, m.UnionID.String)
	}

	return strList
}
