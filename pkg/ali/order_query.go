package ali

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/smartwalle/alipay"
)

type OrderQueryResult struct {
	PaymentStatus string      `json:"paymentStatus"`
	TotalAmount   string      `json:"totalAmount"`
	PaidAmount    string      `json:"paidAmount"`
	InvoiceAmount string      `json:"invoiceAmount"`
	ReceiptAmount string      `json:"receiptAmount"`
	TransactionID string      `json:"transactionId"`
	FtcOrderID    string      `json:"ftcOrderId"`
	PaidAt        chrono.Time `json:"paidAt"`
}

func NewOrderQueryResult(resp *alipay.AliPayTradeQueryResponse) OrderQueryResult {
	return OrderQueryResult{
		PaymentStatus: resp.AliPayTradeQuery.TradeStatus,
		TotalAmount:   resp.AliPayTradeQuery.TotalAmount,
		PaidAmount:    resp.AliPayTradeQuery.BuyerPayAmount,
		InvoiceAmount: resp.AliPayTradeQuery.InvoiceAmount,
		ReceiptAmount: resp.AliPayTradeQuery.ReceiptAmount,
		TransactionID: resp.AliPayTradeQuery.TradeNo,
		FtcOrderID:    resp.AliPayTradeQuery.OutTradeNo,
		PaidAt:        chrono.TimeNow(),
	}
}
