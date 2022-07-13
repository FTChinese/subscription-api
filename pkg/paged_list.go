package pkg

import "github.com/FTChinese/go-rest"

// PagedList is used as the bases to show a list of items with pagination support.
type PagedList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Err error `json:"-"` // Deprecated
}
