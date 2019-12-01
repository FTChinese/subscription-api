package iaprepo

import (
	"database/sql"
	"fmt"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
)

func (env IAPEnv) RetrieveMembership(id reader.MemberID) (subscription.Membership, error) {
	var m subscription.Membership

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
