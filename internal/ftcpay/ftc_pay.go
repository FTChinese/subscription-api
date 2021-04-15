package ftcpay

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/internal/repository/addons"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"go.uber.org/zap"
)

// FtcPay wraps functionalities for user data, subscription and email.
type FtcPay struct {
	SubsRepo     subrepo.Env
	AddOnRepo    addons.Env
	AliPayClient subrepo.AliPayClient
	WxPayClients subrepo.WxPayClientStore
	Postman      postoffice.PostOffice
	Logger       *zap.Logger
}

func New(dbs db.ReadWriteSplit, p postoffice.PostOffice, logger *zap.Logger) FtcPay {

	aliApp := ali.MustInitApp()
	wxApps := wechat.MustGetPayApps()

	return FtcPay{
		SubsRepo:     subrepo.NewEnv(dbs, logger),
		AddOnRepo:    addons.NewEnv(dbs, logger),
		AliPayClient: subrepo.NewAliPayClient(aliApp, logger),
		WxPayClients: subrepo.NewWxClientStore(wxApps, logger),
		Postman:      p,
		Logger:       logger,
	}
}

// SendConfirmEmail sends an email to user after an order is confirmed.
func (pay FtcPay) SendConfirmEmail(pc subs.ConfirmationResult) error {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !pc.Order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := pay.SubsRepo.BaseAccountByUUID(pc.Order.FtcID.String)

	if err != nil {
		return err
	}

	parcel, err := letter.NewSubParcel(account, pc)

	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Info("Send subscription confirmation letter")

	err = pay.Postman.Deliver(parcel)
	if err != nil {
		sugar.Error(err)
		return err
	}
	return nil
}

// ConfirmOrder confirms an order, update membership, backup previous
// membership state, and send email.
func (pay FtcPay) ConfirmOrder(result subs.PaymentResult, order subs.Order) (subs.ConfirmationResult, *subs.ConfirmError) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	sugar.Info("Validate payment result")
	if err := order.ValidatePayment(result); err != nil {
		sugar.Error(err)
		return subs.ConfirmationResult{}, result.ConfirmError(err.Error(), false)
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
		if !confirmed.Snapshot.IsZero() {
			err := pay.SubsRepo.ArchiveMember(confirmed.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

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
func (pay FtcPay) VerifyOrder(order subs.Order) (subs.PaymentResult, error) {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	var payResult subs.PaymentResult
	var err error

	switch order.PaymentMethod {
	case enum.PayMethodWx:
		payResult, err = pay.WxPayClients.VerifyPayment(order)

	case enum.PayMethodAli:
		payResult, err = pay.AliPayClient.VerifyPayment(order)
	}

	if err != nil {
		sugar.Error(err)
		return subs.PaymentResult{}, err
	}

	return payResult, nil
}
