package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/authrepo"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"go.uber.org/zap"
	"net/http"
)

type AuthRouter struct {
	repo    authrepo.Env
	client  ztsms.Client
	postman postoffice.PostOffice
	logger  *zap.Logger
}

func NewAuthRouter(myDBs db.ReadWriteSplit, postman postoffice.PostOffice, l *zap.Logger) AuthRouter {
	return AuthRouter{
		repo:    authrepo.New(myDBs, l),
		client:  ztsms.NewClient(),
		postman: postman,
		logger:  l,
	}
}

// RequestPhoneCode sends a SMS to user.
// Input:
// mobile: string
func (router AuthRouter) RequestPhoneCode(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params ztsms.VerifierParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.ValidateMobile(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf := ztsms.NewVerifier(params.Mobile)

	err := router.repo.SaveSMSVerifier(vrf)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_, err = router.client.SendVerifier(vrf)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	_ = render.New(w).NoContent()
}

// VerifyPhoneCode verifies a code sent to user mobile devices.
// Input:
// mobile: string
// code: string
// deviceToken: string - only required for Android devices.
func (router AuthRouter) VerifyPhoneCode(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params ztsms.VerifierParams

	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	vrf, err := router.repo.RetrieveSMSVerifier(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !vrf.Valid() {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Verification code expired",
			Field:   "code",
			Code:    render.CodeInvalid,
		})
		return
	}

	userID, err := router.repo.UserIDByPhone(vrf.Mobile)
	if err != nil {
		sugar.Error(err)
	}

	go func() {
		err := router.repo.SMSVerifierUsed(vrf.WithUsed())
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(account.NewSearchResult(userID))
}
