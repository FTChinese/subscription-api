package account

import "github.com/guregu/null"

type SearchResult struct {
	ID null.String `json:"id"`
}

func NewSearchResult(id string) SearchResult {
	return SearchResult{
		ID: null.NewString(id, id != ""),
	}
}
