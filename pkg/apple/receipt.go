package apple

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"strings"
)

type ReceiptInput struct {
	ReceiptData string `json:"receiptData"`
}

func (i *ReceiptInput) Validate() *render.ValidationError {
	i.ReceiptData = strings.TrimSpace(i.ReceiptData)

	return validator.New("receiptData").Required().Validate(i.ReceiptData)
}
