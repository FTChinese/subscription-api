package legal

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg/codec"
	"github.com/FTChinese/subscription-api/pkg/conv"
	"github.com/guregu/null"
)

type ContentParams struct {
	Author  string      `json:"author" db:"author"`
	TitleEn string      `json:"titleEn" db:"title_en"`
	TitleCn string      `json:"titleCn" db:"title_cn"`
	Summary null.String `json:"summary" db:"summary"`
	Body    string      `json:"body" db:"body"`
	Keyword null.String `json:"keyword" db:"keyword"`
}

func (p *ContentParams) Validate() *render.ValidationError {
	ve := validator.New("titleEn").Required().Validate(p.TitleEn)
	if ve != nil {
		return ve
	}

	ve = validator.New("titleCn").Required().Validate(p.TitleCn)
	if ve != nil {
		return ve
	}

	return validator.New("body").Required().Validate(p.Body)
}

type Legal struct {
	ID string `json:"id" db:"title_hash"`
	ContentParams
	CreatedUTC chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC chrono.Time `json:"updatedUtc" db:"updated_utc"`
}

func NewLegal(p ContentParams) Legal {
	titleEn := conv.SlashConcat(p.TitleEn)
	hash := codec.HexStringSum(titleEn)
	return Legal{
		ID: hash,
		ContentParams: ContentParams{
			Author:  p.Author,
			TitleEn: titleEn,
			TitleCn: p.TitleCn,
			Summary: p.Summary,
			Body:    p.Body,
			Keyword: p.Keyword,
		},
		CreatedUTC: chrono.TimeNow(),
	}
}

func (l Legal) Update(p ContentParams) Legal {
	l.ContentParams = p

	return l
}

type Teaser struct {
	ID      string      `json:"id" db:"title_hash"`
	TitleEn string      `json:"titleEn" db:"title_en"`
	Summary null.String `json:"summary" db:"summary"`
}

type List struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Legal `json:"data"`
	Err  error   `json:"-"`
}
