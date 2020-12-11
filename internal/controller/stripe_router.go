package controller

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/internal/repository/striperepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/jmoiron/sqlx"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"net/http"
)

type StripeRouter struct {
	config     config.BuildConfig
	signingKey string
	readerRepo readerrepo.Env
	stripeRepo striperepo.Env
	client     striperepo.Client
	logger     *zap.Logger
}

// NewStripeRouter initializes StripeRouter.
func NewStripeRouter(db *sqlx.DB, cfg config.BuildConfig, logger *zap.Logger) StripeRouter {
	client := striperepo.NewClient(cfg.Live(), logger)

	return StripeRouter{
		config:     cfg,
		signingKey: config.MustLoadStripeSigningKey().Pick(cfg.Live()),
		readerRepo: readerrepo.NewEnv(db),
		stripeRepo: striperepo.NewEnv(db, client, logger),
		client:     client,
		logger:     logger,
	}
}

// Forward stripe error to client, and give the error back to caller to handle if it is not stripe error.
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
		err := router.readerRepo.ArchiveMember(result.Snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}
}

// IssueKey creates an ephemeral key.
//
// POST /stripe/customers/{id}/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get stripe customer id.
	cusID, err := getURLParam(req, "id").ToString()
	if err != nil {
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
		err = handleErrResp(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_, err = w.Write(keyData)
	if err != nil {
		sugar.Error(err)
	}
}
