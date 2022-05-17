package legal

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
)

type ContentParams struct {
	Author  string      `json:"author" db:"author"`
	Title   string      `json:"title" db:"title"`
	Summary null.String `json:"summary" db:"summary"`
	Body    string      `json:"body" db:"body"`
	Keyword null.String `json:"keyword" db:"keyword"`
}

func (p *ContentParams) Validate() *render.ValidationError {

	ve := validator.New("title").Required().Validate(p.Title)
	if ve != nil {
		return ve
	}

	return validator.New("body").Required().Validate(p.Body)
}

type Legal struct {
	HashID string `json:"id" db:"hash_id"`
	ContentParams
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC chrono.Time `json:"updatedUtc" db:"updated_utc"`
}

func NewLegal(p ContentParams) Legal {

	id := ids.LegalDocID()
	return Legal{
		HashID:        id,
		ContentParams: p,
		CreatedUTC:    chrono.TimeNow(),
	}
}

func (l Legal) Update(p ContentParams) Legal {
	l.ContentParams = p

	return l
}

type Teaser struct {
	ID      string      `json:"id" db:"hash_id"`
	Title   string      `json:"title" db:"title"`
	Summary null.String `json:"summary" db:"summary"`
}

type List struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Legal `json:"data"`
	Err  error   `json:"-"`
}
