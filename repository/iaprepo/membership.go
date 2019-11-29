package iaprepo

import (
	"database/sql"
	"fmt"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/reader"
)

func (env IAPEnv) RetrieveMembership(id reader.MemberID) (paywall.Membership, error) {
	var m paywall.Membership

	err := env.db.Get(
		&m,
		fmt.Sprintf(selectMember, id.MemberColumn()),
		id.CompoundID,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.WithField("trace", "IAPEnv.RetrieveMember").Error(err)
		return m, err
	}

	m.Normalize()

	return m, nil
}

func CreateMembership() error {
	return nil
}
