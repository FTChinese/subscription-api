package shared

import (
	"github.com/FTChinese/subscription-api/pkg/db"
)

// ReaderBaseRepo contains shared functionalities of a reader.
// It contains methods to retrieve user's
// basic account data using various id fields.
// It also contains methods to retrieve membership
// using various ids.
type ReaderBaseRepo struct {
	DBs db.ReadWriteMyDBs
}

func New(dbs db.ReadWriteMyDBs) ReaderBaseRepo {
	return ReaderBaseRepo{
		DBs: dbs,
	}
}
