package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/stripeenv"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"go.uber.org/zap"
	"net/http"
)

type StripeRouter struct {
	SigningKey     string
	PublishableKey string
	Env            stripeenv.Env
	ReaderRepo     shared.ReaderCommon
	Logger         *zap.Logger
	Live           bool
}

func (router StripeRouter) handleSubsResult(result stripe.SubsResult) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	err := router.Env.UpsertSubs(result.Subs, true)
	if err != nil {
		sugar.Error(err)
	}

	err = router.Env.UpsertInvoice(result.Subs.LatestInvoice)
	if err != nil {
		sugar.Error(err)
	}

	err = router.Env.UpsertPaymentIntent(result.Subs.PaymentIntent)
	if err != nil {
		sugar.Error(err)
	}

	if !result.Versioned.IsZero() {
		err := router.ReaderRepo.VersionMembership(result.Versioned)
		if err != nil {
			sugar.Error(err)
		}
	}
}

func handleSubsError(w http.ResponseWriter, err error) {
	switch err {
	case reader.ErrTrialUpgradeForbidden:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "trial_upgrade",
			Code:    render.CodeInvalid,
		})

	case reader.ErrAlreadyStripeSubs,
		reader.ErrAlreadyAppleSubs,
		reader.ErrAlreadyB2BSubs:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "membership",
			Code:    render.CodeAlreadyExists,
		})

	case reader.ErrUnknownPaymentMethod:
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: err.Error(),
			Field:   "payMethod",
			Code:    render.CodeInvalid,
		})

	default:
		_ = xhttp.HandleStripeErr(w, err)
	}
}
