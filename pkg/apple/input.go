package apple

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"strings"
)

type ReceiptInput struct {
	ReceiptData       string `json:"receiptData"`
	LegacyReceiptData string `json:"receipt-data"`
}

func (i *ReceiptInput) Validate() *render.ValidationError {
	i.ReceiptData = strings.TrimSpace(i.ReceiptData)
	i.LegacyReceiptData = strings.TrimSpace(i.LegacyReceiptData)

	return validator.New("receiptData").Required().Validate(i.ReceiptData)
}

// LinkInput defines the request body to link IAP to ftc account.
type LinkInput struct {
	FtcID        string `json:"ftcId"`
	OriginalTxID string `json:"originalTxId"`
}

func (i *LinkInput) Validate() *render.ValidationError {
	i.FtcID = strings.TrimSpace(i.FtcID)
	i.OriginalTxID = strings.TrimSpace(i.OriginalTxID)

	ve := validator.New("ftcId").Required().Validate(i.FtcID)
	if ve != nil {
		return ve
	}

	return validator.New("originalTxId").Required().Validate(i.OriginalTxID)
}
