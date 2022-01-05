package ali

import (
	"fmt"
	"github.com/smartwalle/alipay"
	"go.uber.org/zap"
	"net/http"
)

type PayClient struct {
	app    App
	sdk    *alipay.AliPay
	logger *zap.Logger
}

func NewPayClient(app App, logger *zap.Logger) PayClient {
	return PayClient{
		app:    app,
		sdk:    alipay.New(app.ID, app.PublicKey, app.PrivateKey, true),
		logger: logger,
	}
}

// CreateOrder signs order creation param and returns the signed string as query parameters.
// Inside browsers, the query parameter is actually an url you can redirect to.
func (c PayClient) CreateOrder(or OrderReq) (string, error) {
	param := alipay.TradePay{
		NotifyURL:   or.WebhookURL,
		Subject:     or.Title,
		OutTradeNo:  or.FtcOrderID,
		TotalAmount: or.TotalAmount,
		GoodsType:   "0",
	}

	switch or.TxKind {
	case EntryApp:
		param.ProductCode = string(ProductCodeApp)
		return c.sdk.TradeAppPay(alipay.AliPayTradeAppPay{
			TradePay: param,
		})

	case EntryDesktopWeb:
		param.ProductCode = string(ProductCodeWeb)
		param.ReturnURL = or.ReturnURL
		url, err := c.sdk.TradePagePay(alipay.AliPayTradePagePay{
			TradePay: param,
		})
		if err != nil {
			return "", err
		}

		return url.String(), nil

	case EntryMobileWeb:
		param.ProductCode = string(ProductCodeWeb)
		param.ReturnURL = or.ReturnURL
		url, err := c.sdk.TradeWapPay(alipay.AliPayTradeWapPay{
			TradePay: param,
		})
		if err != nil {
			return "", err
		}

		return url.String(), nil
	}

	return "", fmt.Errorf("unknown payment platform %s", or.TxKind)
}

// QueryOrder retrieves an order from alipay api.
// https://opendocs.alipay.com/apis/api_1/alipay.trade.query/
func (c PayClient) QueryOrder(id string) (*alipay.AliPayTradeQueryResponse, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	qr, err := c.sdk.TradeQuery(alipay.AliPayTradeQuery{
		AppAuthToken: "",
		OutTradeNo:   id,
	})

	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	sugar.Infof("Alipay order query result: %v", qr)

	// Check if the http request is successfully made.
	// Example failure response:
	// Code:40004  // 网关返回码
	// Msg:Business Failed // 网关返回码描述
	// SubCode:ACQ.TRADE_NOT_EXIST // 业务返回码
	// SubMsg:交易不存在 // 业务返回码描述
	// AuthTradePayMode:
	// BuyerLogonId: BuyerPayAmount:0.00
	// BuyerUserId:
	// BuyerUserType:
	// InvoiceAmount:0.00
	// OutTradeNo:FT33381C4D23AE4F19
	// PointAmount:0.00 ReceiptAmount:0.00
	// SendPayDate:
	// TotalAmount:
	// TradeNo:
	// TradeStatus: // 交易状态：WAIT_BUYER_PAY（交易创建，等待买家付款）、TRADE_CLOSED（未付款交易超时关闭，或支付完成后全额退款）、TRADE_SUCCESS（交易支付成功）、TRADE_FINISHED（交易结束，不可退款）
	// DiscountAmount:
	// FundBillList:[]
	// MdiscountAmount:
	// PayAmount:
	// PayCurrency:
	// SettleAmount:
	// SettleCurrency:
	// SettleTransRate:
	// StoreId: StoreName:
	// TerminalId: TransCurrency:
	// TransPayRate:
	// DiscountGoodsDetail:
	// IndustrySepcDetail:
	// VoucherDetailList:[]
	if !qr.IsSuccess() {
		sugar.Infof("Alipay query order failed %s, %s", qr.AliPayTradeQuery.Msg, qr.AliPayTradeQuery.SubCode)

		return nil, fmt.Errorf("failure calling alipay api: %s - %s", qr.AliPayTradeQuery.Code, qr.AliPayTradeQuery.Msg)
	}

	return qr, nil
}

// GetWebhookPayload retrieves wehbook data from request and verify if it targeted at us.
func (c PayClient) GetWebhookPayload(req *http.Request) (*alipay.TradeNotification, error) {
	payload, err := c.sdk.GetTradeNotification(req)
	if err != nil {
		return nil, err
	}

	// 验证app_id是否为该商户本身
	matched := c.app.ID == payload.AppId
	if !matched {
		return payload, fmt.Errorf("mismatched ali app id, expected %s, got %s", c.app.ID, payload.AppId)
	}

	return payload, nil
}
