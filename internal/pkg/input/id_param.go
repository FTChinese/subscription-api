package input

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
)

type IDParam struct {
	ID string `json:"id"`
}

func (p IDParam) Validate() *render.ValidationError {
	return validator.New("id").Required().Validate(p.ID)
}
