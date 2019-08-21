package repository

import "gitlab.com/ftchinese/subscription-api/models/paywall"

// AddMemberID set a membership's id column if it is empty.
func (env Env) AddMemberID(m paywall.Membership) error {
	_, err := env.db.NamedExec(
		env.query.AddMemberID(m.MemberColumn()),
		m)

	if err != nil {
		return err
	}

	return nil
}
