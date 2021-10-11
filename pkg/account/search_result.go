package account

import "github.com/guregu/null"

// SearchResult contains a user's uuid to indicates whether
// the user is found.
type SearchResult struct {
	ID null.String `json:"id" db:"id"`
}

func NewSearchResult(id string) SearchResult {
	return SearchResult{
		ID: null.NewString(id, id != ""),
	}
}
