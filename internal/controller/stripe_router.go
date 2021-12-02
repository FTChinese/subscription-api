package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/addons"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/internal/repository/striperepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"net/http"
)

type StripeRouter struct {
	signingKey string
	addOnRepo  addons.Env
	stripeRepo striperepo.Env
	client     stripeclient.Client
	logger     *zap.Logger
	isLive     bool
}

// NewStripeRouter initializes StripeRouter.
func NewStripeRouter(
	readerBase shared.ReaderBaseRepo,
	stripeBase shared.StripeBaseRepo,
	dbs db.ReadWriteMyDBs,
	logger *zap.Logger,
	isLive bool,
) StripeRouter {
	client := stripeclient.New(isLive, logger)

	return StripeRouter{
		signingKey: config.MustStripeWebhookKey().Pick(isLive),
		addOnRepo:  addons.NewEnv(dbs, logger),
		stripeRepo: striperepo.New(readerBase, stripeBase, logger),
		client:     client,
		logger:     logger,
	}
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	err := router.stripeRepo.UpsertSubs(result.Subs)
	if err != nil {
		sugar.Error(err)
	}

	if !result.Snapshot.IsZero() {
		err := router.stripeRepo.ArchiveMember(result.Snapshot)
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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

	keyData, err := router.client.CreateEphemeralKey(cusID, stripeVersion)
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
