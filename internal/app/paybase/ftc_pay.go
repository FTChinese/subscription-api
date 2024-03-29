package paybase

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/internal/pkg/letter"
	"github.com/FTChinese/subscription-api/internal/repository/addons"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"go.uber.org/zap"
)

// FtcPayBase wraps shared functionalities used for both api's one-time pay and polling service.
type FtcPayBase struct {
	SubsRepo     subrepo.Env
	ReaderRepo   shared.ReaderCommon
	AddOnRepo    addons.Env
	AliPayClient ali.PayClient
	WxPayClients wechat.WxPayClientStore
	EmailService letter.Service
	Logger       *zap.Logger
}

func NewFtcPay(
	dbs db.ReadWriteMyDBs,
	logger *zap.Logger,
) FtcPayBase {
	return FtcPayBase{
		SubsRepo:     subrepo.New(dbs, logger),
		ReaderRepo:   shared.NewReaderCommon(dbs),
		AddOnRepo:    addons.New(dbs, logger),
		AliPayClient: ali.NewPayClient(ali.MustInitApp(), logger),
		WxPayClients: wechat.NewWxClientStore(wechat.MustGetPayApps(), logger),
		EmailService: letter.NewService(logger),
		Logger:       logger,
	}
}

// SendConfirmEmail sends an email to user after an order is confirmed.
func (pay FtcPayBase) SendConfirmEmail(result ftcpay.ConfirmationResult) error {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !result.Order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := pay.ReaderRepo.BaseAccountByUUID(result.Order.FtcID.String)

	if err != nil {
		return err
	}

	err = pay.EmailService.SendOneTimePurchase(account, result.Invoices)
	if err != nil {
		sugar.Error(err)
		return err
	}
	return nil
}

// ConfirmOrder confirms an order, update membership, backup previous
// membership state, and send email.
// Used by both webhook and client verification.
func (pay FtcPayBase) ConfirmOrder(result ftcpay.PaymentResult, order ftcpay.Order) (ftcpay.ConfirmationResult, *ftcpay.ConfirmError) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	sugar.Info("Validate payment result")
	if err := order.ValidatePayment(result); err != nil {
		sugar.Error(err)
		return ftcpay.ConfirmationResult{}, result.ConfirmError(err.Error(), false)
	}

	confirmed, cfmErr := pay.SubsRepo.ConfirmOrder(result, order)
	if cfmErr != nil {
		go func() {
			err := pay.SubsRepo.SaveConfirmErr(cfmErr)
			if err != nil {
				sugar.Error(err)
			}
		}()
		return confirmed, cfmErr
	}

	go func() {
		// Save membership change history.
		if !confirmed.Versioned.IsZero() {
			err := pay.ReaderRepo.VersionMembership(confirmed.Versioned)
			if err != nil {
				sugar.Error(err)
			}
		}

		// Flag invoices as carried over if there are
		// unused portion.
		if !confirmed.Invoices.CarriedOver.IsZero() {
			err := pay.AddOnRepo.InvoicesCarriedOver(confirmed.Membership.UserIDs)
			if err != nil {
				sugar.Error(err)
			}
		}

		if confirmed.Notify {
			err := pay.SendConfirmEmail(confirmed)
			if err != nil {
				sugar.Error(err)
			}
		}
	}()

	return confirmed, nil
}

// VerifyOrder verifies against payment providers that an order is actually paid.
func (pay FtcPayBase) VerifyOrder(order ftcpay.Order) (ftcpay.PaymentResult, error) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	var payResult ftcpay.PaymentResult
	var err error

	switch order.PaymentMethod {
	case enum.PayMethodWx:
		payResult, err = pay.verifyWxOrder(order)

	case enum.PayMethodAli:
		payResult, err = pay.verifyAliOrder(order)
	}

	if err != nil {
		sugar.Error(err)
		return ftcpay.PaymentResult{}, err
	}

	return payResult, nil
}

func (pay FtcPayBase) verifyAliOrder(order ftcpay.Order) (ftcpay.PaymentResult, error) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	aliOrder, err := pay.AliPayClient.QueryOrder(order.ID)
	if err != nil {
		sugar.Error(err)
		return ftcpay.PaymentResult{}, err
	}

	sugar.Infof("Alipay raw order: %+v", aliOrder)

	go func() {
		err := pay.SubsRepo.SaveAliOrderQueryPayload(
			ali.NewOrderQueryPayload(aliOrder))

		if err != nil {
			sugar.Error(err)
		}
	}()

	return ftcpay.NewAliPayResult(aliOrder), nil
}

func (pay FtcPayBase) verifyWxOrder(order ftcpay.Order) (ftcpay.PaymentResult, error) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	client, err := pay.WxPayClients.FindByAppID(order.WxAppID.String)
	if err != nil {
		sugar.Error(err)
		return ftcpay.PaymentResult{}, err
	}

	payload, err := client.QueryOrder(wechat.NewOrderQueryParams(order.ID))
	if err != nil {
		sugar.Error(err)
		return ftcpay.PaymentResult{}, err
	}

	sugar.Infof("Wxpay raw order %+v", payload)

	go func() {
		err := pay.SubsRepo.SaveWxPayload(
			wechat.NewPayloadSchema(
				order.ID,
				payload,
			).WithKind(wechat.RowKindQueryOrder),
		)
		if err != nil {
			sugar.Error(err)
		}
	}()

	return ftcpay.NewWxPayResult(wechat.NewOrderQueryResp(payload)), nil
}
