package ftcpay

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/internal/repository/subrepo"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// FtcPay wraps functionalities for user data, subscription and email.
type FtcPay struct {
	SubsRepo     subrepo.Env
	ReaderRepo   readerrepo.Env
	AliPayClient subrepo.AliPayClient
	WxPayClients subrepo.WxPayClientStore
	Postman      postoffice.PostOffice
	Logger       *zap.Logger
}

func New(db *sqlx.DB, p postoffice.PostOffice, logger *zap.Logger) FtcPay {

	aliApp := ali.MustInitApp()
	wxApps := wechat.MustGetPayApps()

	return FtcPay{
		SubsRepo:     subrepo.NewEnv(db, logger),
		ReaderRepo:   readerrepo.NewEnv(db),
		AliPayClient: subrepo.NewAliPayClient(aliApp, logger),
		WxPayClients: subrepo.NewWxClientStore(wxApps, logger),
		Postman:      p,
		Logger:       logger,
	}
}

// SendConfirmEmail sends an email to user after an order is confirmed.
func (pay FtcPay) SendConfirmEmail(order subs.Order) error {
	defer pay.Logger.Sync()
	sugar := pay.Logger.Sugar()

	// If the FtcID field is null, it indicates this user
	// does not have an FTC account linked. You cannot find out
	// its email address.
	if !order.FtcID.Valid {
		return nil
	}
	// Find this user's personal data
	account, err := pay.ReaderRepo.AccountByFtcID(order.FtcID.String)

	if err != nil {
		return err
	}

	var parcel postoffice.Parcel
	switch order.Kind {
	case enum.OrderKindCreate:
		parcel, err = letter.NewSubParcel(account, order)

	case enum.OrderKindRenew:
		parcel, err = letter.NewRenewalParcel(account, order)

	case enum.OrderKindUpgrade:
		pos, err := pay.SubsRepo.ListProratedOrders(order.ID)
		if err != nil {
			return err
		}
		parcel, err = letter.NewUpgradeParcel(account, order, pos)
	}

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
			err := pay.ReaderRepo.ArchiveMember(confirmed.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

		// Flag upgrade balance as consumed.
		if confirmed.Order.Kind == enum.OrderKindUpgrade {
			err := pay.SubsRepo.ProratedOrdersUsed(confirmed.Order.ID)
			if err != nil {
				sugar.Error(err)
			}
		}

		if err := pay.SendConfirmEmail(confirmed.Order); err != nil {
			sugar.Error(err)
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
