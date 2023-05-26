package pkg

import gorest "github.com/FTChinese/go-rest"

// PagedList is used as the bases to show a list of items with pagination support.
type PagedData[T interface{}] struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []T `json:"data"`
}
