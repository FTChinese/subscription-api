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
	SigningKey string
	Env        stripeenv.Env
	ReaderRepo shared.ReaderCommon
	Logger     *zap.Logger
	Live       bool
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

	if !result.Snapshot.IsZero() {
		err := router.ReaderRepo.ArchiveMember(result.Snapshot)
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

// IssueKey creates an ephemeral key.
// https://stripe.com/docs/mobile/android/basic#set-up-ephemeral-key
//
// POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get stripe customer id.
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	stripeVersion := req.FormValue("api_version")
	if stripeVersion == "" {
		_ = render.New(w).BadRequest("Stripe-Version not found")
		return
	}

	keyData, err := router.Env.Client.CreateEphemeralKey(cusID, stripeVersion)
	if err != nil {
		sugar.Error(err)
		err = xhttp.HandleStripeErr(w, err)
		if err == nil {
			return
		}
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	_, err = w.Write(keyData)
	if err != nil {
		sugar.Error(err)
	}
}
