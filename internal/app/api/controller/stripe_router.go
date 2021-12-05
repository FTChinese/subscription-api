package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/internal/repository/striperepo"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"net/http"
)

type StripeRouter struct {
	SigningKey      string
	StripeRepo      striperepo.Env
	StripePriceRepo shared.StripeBaseRepo
	ReaderRepo      shared.ReaderCommon
	Client          stripeclient.Client
	Logger          *zap.Logger
	Live            bool
}

// Forward stripe error to smsClient, and give the error back to caller to handle if it is not stripe error.
func handleErrResp(w http.ResponseWriter, err error) error {

	var se *stripeSdk.Error
	var ve *render.ValidationError
	var re *render.ResponseError
	switch {
	case errors.As(err, &se):
		return render.New(w).JSON(se.HTTPStatusCode, se)

	case errors.As(err, &ve):
		return render.New(w).Unprocessable(ve)

	case errors.As(err, &re):
		return render.New(w).JSON(re.StatusCode, re)

	default:
		return err
	}
}

func (router StripeRouter) handleSubsResult(result stripe.SubsResult) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	err := router.StripeRepo.UpsertSubs(result.Subs)
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

// IssueKey creates an ephemeral key.
// https://stripe.com/docs/mobile/android/basic#set-up-ephemeral-key
//
// POST /stripe/customers/{id}/ephemeral-keys?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	// Get stripe customer id.
	cusID, err := getURLParam(req, "id").ToString()
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

	keyData, err := router.Client.CreateEphemeralKey(cusID, stripeVersion)
	if err != nil {
		sugar.Error(err)
		err = handleErrResp(w, err)
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
