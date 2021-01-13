package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/guregu/null"
	"github.com/smartwalle/alipay"
	"strconv"
)

func priceToCent(s string) null.Int {
	if s == "" {
		return null.Int{}
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return null.Int{}
	}

	return null.IntFrom(int64(f * 100))
}

const colSavePayResult = `
payment_status = :payment_status,
	status_detail = :status_detail,
	paid_amount = :paid_amount,
	tx_id = :tx_id,
	updated_utc = UTC_TIMESTAMP()
`

const StmtSavePayResult = `
INSERT INTO premium.ftc_pay_result
SET order_id = :order_id,` +
	colSavePayResult + `,
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
` + colSavePayResult

// PaymentResult unifies ali and wx webhook payload, or query order.
// TRADE_FINISHED：交易成功且结束，即不可再做任何操作
// 例如在高级即时到帐接口里面，支付成功之后返回的是TRADE_SUCCESS，此时三个月之内可以操作退款，三个月之后不允许对该笔交易操作，支付宝会返回TRADE_FINISHED，所以必须要在TRADE_SUCCESS下执行你网站业务逻辑代码，TRADE_FINISHED不做任何业务逻辑处理，避免一笔交易重复执行业务逻辑而给您带来不必要的损失。
type PaymentResult struct {
	// // 在支付宝的业务通知中，只有交易通知状态为TRADE_SUCCESS或TRADE_FINISHED时，支付宝才会认定为买家付款成功。
	// For Alipay `trade_status` field:
	// WAIT_BUYER_PAY（交易创建，等待买家付款）
	// TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）
	// TRADE_SUCCESS（交易支付成功）
	// TRADE_FINISHED（交易结束，不可退款）
	// For Wechat `trade_state` field:
	// SUCCESS—支付成功
	// REFUND—转入退款
	// NOTPAY—未支付
	// CLOSED—已关闭
	// REVOKED—已撤销（刷卡支付）
	// USERPAYING--用户支付中
	// PAYERROR--支付失败(其他原因，如银行返回失败)
	PaymentState string `json:"paymentState" db:"payment_status"`
	// For wechat `trade_state_desc` field.
	PaymentStateDesc string `json:"paymentStateDesc" db:"status_detail"`
	// In cent.
	// For alipay, we use the total_amount, which is
	// 交易的订单金额，单位为元，两位小数。该参数的值为支付时传入的total_amount.
	// For our purpose, that is the amount we actually charged user.
	Amount        null.Int       `json:"totalFee" db:"paid_amount"` // In cent
	TransactionID string         `json:"transactionId" db:"tx_id"`
	OrderID       string         `json:"ftcOrderId" db:"order_id"`
	PaidAt        chrono.Time    `json:"paidAt" db:"paid_at"`
	ConfirmedUTC  chrono.Time    `json:"-"` // Use this an order's confirmation time. For webhook this is the same as PaidAt. For order query, they are different.
	PayMethod     enum.PayMethod `json:"payMethod"`
}

// ShouldRetry checks if we should tell webhook should resend notification.
func (r PaymentResult) ShouldRetry() bool {
	switch r.PaymentState {
	case ali.TradeStatusPending, wechat.TradeStatePaying:
		return true

	default:
		return false
	}
}

// IsOrderPaid checks if a wx or ali order is paid.
// Do not call it for webhook since the payload does not contain these fields.
func (r PaymentResult) IsOrderPaid() bool {
	switch r.PaymentState {
	case ali.TradeStatusSuccess, wechat.TradeStateSuccess:
		return true
	}

	return false
}

func (r PaymentResult) ConfirmError(msg string, retry bool) *ConfirmError {
	return &ConfirmError{
		OrderID: r.OrderID,
		Message: msg,
		Retry:   retry,
	}
}

// NewWxWebhookResult builds PaymentResult from wechat pay webhook notification.
func NewWxWebhookResult(payload wechat.Notification) PaymentResult {
	paidAt := chrono.TimeFrom(dt.MustParseWxTime(payload.TimeEnd.String))

	return PaymentResult{
		PaymentState:     payload.ResultCode.String,
		PaymentStateDesc: payload.ErrorMessage.String,
		Amount:           payload.TotalFee,
		TransactionID:    payload.TransactionID.String,
		OrderID:          payload.FTCOrderID.String,
		PaidAt:           paidAt,
		ConfirmedUTC:     paidAt,
		PayMethod:        enum.PayMethodWx,
	}
}

// NewWxPayResult creates a new PaymentResult from the result of querying wechat order.
// The payment status should be returned to client as is, whether it is paid or not.
func NewWxPayResult(r wechat.OrderQueryResp) PaymentResult {

	paidAt, _ := dt.ParseWxTime(r.TimeEnd.String)

	pr := PaymentResult{
		PaymentState:     r.TradeState.String,
		PaymentStateDesc: r.TradeStateDesc.String,
		Amount:           null.IntFrom(r.TotalFee.Int64),
		TransactionID:    r.TransactionID.String,
		OrderID:          r.FTCOrderID.String,
		PaidAt:           chrono.TimeFrom(paidAt),
		PayMethod:        enum.PayMethodWx,
	}

	if pr.IsOrderPaid() {
		pr.ConfirmedUTC = chrono.TimeNow()
	}

	return pr
}

// NewAliPayResult converts ali payment result.
// trade_status
// total_amount: 交易的订单金额，单位为元，两位小数。该参数的值为支付时传入的total_amount
// settle_amount：结算币种订单金额
// pay_amount： 支付币种订单金额
// buyer_pay_amount： 买家实付金额，单位为元，两位小数。该金额代表该笔交易买家实际支付的金额，不包含商户折扣等金额
// invoice_amount： 交易中用户支付的可开具发票的金额，单位为元，两位小数
// receipt_amount： 实收金额，单位为元，两位小数。该金额为本笔交易，商户账户能够实际收到的金额
// An example:
// &{AliPayTradeQuery:{
// Code:10000
// Msg:Success
// SubCode:
// SubMsg:
// AuthTradePayMode:
// BuyerLogonId:niw***@outlook.com
// BuyerPayAmount:0.00
// BuyerUserId:2088102779940036
// BuyerUserType:
// InvoiceAmount:0.00
// OutTradeNo:FT8F0438FFE67C7443
// PointAmount:0.00
// ReceiptAmount:0.00
// SendPayDate:2020-11-03 21:14:43
// TotalAmount:0.01
// TradeNo:2020110322001440031406041664
// TradeStatus:TRADE_SUCCESS
// DiscountAmount:
// FundBillList:[]
// MdiscountAmount:
// PayAmount:
// PayCurrency:
// SettleAmount:
// SettleCurrency:
// SettleTransRate:
// StoreId:
// StoreName:
// TerminalId:
// TransCurrency:
// TransPayRate:
// DiscountGoodsDetail:
// IndustrySepcDetail:
// VoucherDetailList:[]
func NewAliPayResult(r *alipay.AliPayTradeQueryResponse) PaymentResult {
	paidAt, _ := dt.ParseAliTime(r.AliPayTradeQuery.SendPayDate)

	pr := PaymentResult{
		PaymentState:     r.AliPayTradeQuery.TradeStatus,
		PaymentStateDesc: ali.TradeStatusMsg[r.AliPayTradeQuery.TradeStatus],
		Amount:           priceToCent(r.AliPayTradeQuery.TotalAmount),
		TransactionID:    r.AliPayTradeQuery.TradeNo,
		OrderID:          r.AliPayTradeQuery.OutTradeNo,
		PaidAt:           chrono.TimeFrom(paidAt), // If an order is not confirmed, always use now as confirmation time.
		PayMethod:        enum.PayMethodAli,
	}

	if pr.IsOrderPaid() {
		pr.ConfirmedUTC = chrono.TimeNow()
	}

	return pr
}

// NewAliWebhookResult builds PaymentResult from alipay webhook notification.
//	total_amount		VARCHAR(10), 本次交易支付的订单金额，单位为人民币（元）
//	receipt_amount		VARCHAR(10), 商家在交易中实际收到的款项，单位为人民币, 只有TRADE_SUCCESS有，其他为空。
//	invoice_amount		VARCHAR(10), 用户在交易中支付的可开发票的金额
func NewAliWebhookResult(payload *alipay.TradeNotification) (PaymentResult, error) {
	// ReceiptAmount is the actually paid.
	paidAt := chrono.TimeFrom(dt.MustParseAliTime(payload.GmtPayment))
	return PaymentResult{
		PaymentState:     payload.TradeStatus,
		PaymentStateDesc: ali.TradeStatusMsg[payload.TradeStatus],
		Amount:           priceToCent(payload.ReceiptAmount),
		TransactionID:    payload.TradeNo,
		OrderID:          payload.OutTradeNo,
		PaidAt:           paidAt,
		ConfirmedUTC:     paidAt,
		PayMethod:        enum.PayMethodAli,
	}, nil
}
