package account

import "github.com/FTChinese/go-rest/chrono"

const StmtSaveDeletedUser = `
INSERT INTO user_db.deleted_user
SET user_id = :id,
	email = :email,
	created_utc = :created_utc
`

const StmtDeleteUser = `
DELETE from cmstmp01.userinfo
WHERE user_id = ?
LIMIT 1`

const StmtDeleteProfile = `
DELETE FROM user_db.profile
WHERE user_id = ?
LIMIT 1`

type DeletedUser struct {
	ID         string      `db:"id"`
	Email      string      `db:"email"`
	CreatedUTC chrono.Time `db:"created_utc"`
}

func (a BaseAccount) Deleted() DeletedUser {
	return DeletedUser{
		ID:         a.FtcID,
		Email:      a.Email,
		CreatedUTC: chrono.TimeNow(),
	}
}
