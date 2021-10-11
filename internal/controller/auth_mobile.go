package controller

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"net/http"
)

// RequestSMSVerification sends an SMS to user for login.
// Input:
// mobile: string
func (router AuthRouter) RequestSMSVerification(w http.ResponseWriter, req *http.Request) {
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

	// Retrieve account by mobile number.
	// If not found, it indicates the mobile is used for the first time.
	mobileFound, err := router.userRepo.SearchByMobile(params.Mobile)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Fallthrough if mobileFound is zero value.
	// It only indicates user might be logging for the 1st time.

	// Create the verifier. If user id does not exist, it indicates
	// user is using mobile to login for the first t ime.
	vrf := ztsms.NewVerifier(params.Mobile, mobileFound.ID)

	err = router.userRepo.SaveSMSVerifier(vrf)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Send the code to user device.
	_, err = router.smsClient.SendVerifier(vrf)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	_ = render.New(w).NoContent()
}

// VerifySMSCode verifies a code sent to user mobile devices.
// Client must send mobile number and the SMS code together as
// the code is short and collision chances are high.
// When retrieving data, only find those rows whose used_ftc is null,
// expiration time is not past yet.
// The SMS code has only 6 digits, which means the collision rate is
// 1/1000000.
//
// Input:
// * mobile: string - the mobile number used for login
// * code: string - the SMS cod of this session
// * deviceToken?: string; - only required for Android devices.
//
// Required header: footprint.Client
//
// Returns account.SearchResult containing nullable user id.
// If user id is null, it indicates this mobile phone is used for the first time.
// Client should ask user to enter email so that we could link this mobile to an email account;
// otherwise client should use the user id to retrieve reader.Account.
func (router AuthRouter) VerifySMSCode(w http.ResponseWriter, req *http.Request) {
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

	// Retrieve verifier using mobile + code.
	// Not found could be produced if the code does not exist or is expired.
	vrf, err := router.userRepo.RetrieveSMSVerifier(params)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		err := router.userRepo.SMSVerifierUsed(vrf.WithUsed())
		if err != nil {
			sugar.Error(err)
		}
	}()

	// If FtcID exists, it indicates this mobile is already
	// linked to an email account. We treat it as a login
	// session and record client metadata;
	// otherwise the metadata should be recorded by link mobile
	// or signup process.
	if vrf.FtcID.Valid {
		fp := footprint.
			New(vrf.FtcID.String, footprint.NewClient(req)).
			FromLogin().
			WithAuth(enum.LoginMethodMobile, params.DeviceToken)

		go func() {
			err := router.userRepo.SaveFootprint(fp)
			if err != nil {
				sugar.Error(err)
			}
		}()

		_ = render.New(w).OK(account.NewSearchResult(vrf.FtcID.String))
		return
	}

	// If it goes here, it indicates user_db.profile table
	// does not have a row with the mobile number provided in request body.
	// Then we use mobile's faked email to search user in the userinfo table.
	result, err := router.userRepo.SearchByEmail(account.MobileEmail(params.Mobile))
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).OK(account.SearchResult{})
		return
	}

	// If user id is found this way,
	// it indicates the mobile number exists in userinfo table
	// but does not exist in profile table.
	// Two cases here:
	// 1. profile table does not have a row for this ftc id, insert directly.
	// 2. profile table have this ftc id and mobile_phone column is empty, update it.
	if result.ID.Valid {
		go func() {
			err := router.userRepo.UpsertMobile(ztsms.MobileUpdater{
				FtcID:  result.ID.String,
				Mobile: null.StringFrom(params.Mobile),
			})
			if err != nil {
				sugar.Error(err)
				fp := footprint.
					New(result.ID.String, footprint.NewClient(req)).
					FromLogin().
					WithAuth(enum.LoginMethodMobile, params.DeviceToken)

				err := router.userRepo.SaveFootprint(fp)
				if err != nil {
					sugar.Error(err)
				}
			}
		}()
	}

	_ = render.New(w).OK(result)
}

// LinkMobile authenticates an existing email account,
// and link to the provided mobile phone.
//
// Input:
// email: string;
// password: string;
// mobile: string;
// deviceToken?: string;
// sourceUrl?: string; // Used to compose email verification link.
//
// Returns reader.Account.
func (router AuthRouter) LinkMobile(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params input.MobileLinkParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Find the user id and password matching state by email.
	// If not found, it indicates the account does not exist.
	authResult, err := router.userRepo.Authenticate(params.EmailLoginParams)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Account is found, but password does not match.
	if !authResult.PasswordMatched {
		_ = render.New(w).Forbidden("Incorrect credentials")
		return
	}

	// Credentials authenticated, set mobile to this account.
	err = router.userRepo.UpsertMobile(ztsms.MobileUpdater{
		FtcID:  authResult.UserID,
		Mobile: null.StringFrom(params.Mobile),
	})

	if err != nil {
		if errors.Is(err, ztsms.ErrMobileAlreadyExists) {
			// If already set to to other mobile
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: "This email account is already linked to another mobile",
				Field:   "mobile",
				Code:    render.CodeAlreadyExists,
			})
			return
		}
		if errors.Is(err, ztsms.ErrMobileAlreadySet) {
			acnt, err := router.userRepo.AccountByFtcID(authResult.UserID)
			if err != nil {
				// There shouldn't be ErrNoRow error here.
				sugar.Error(err)
				_ = render.New(w).DBError(err)
				return
			}
			_ = render.New(w).OK(acnt)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Retrieve account by user id.
	// There shouldn't be any 404 error here.
	acnt, err := router.userRepo.AccountByFtcID(authResult.UserID)
	if err != nil {
		// There shouldn't be ErrNoRow error here.
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Tracking.
	clientApp := footprint.NewClient(req)
	fp := footprint.New(acnt.FtcID, clientApp).
		FromSignUp().
		WithAuth(enum.LoginMethodMobile, params.DeviceToken)

	go func() {
		err := router.userRepo.SaveFootprint(fp)
		if err != nil {
			sugar.Error()
		}
		// TODO: send email?
	}()

	_ = render.New(w).OK(acnt)
}

// MobileSignUp creates a new mobile account.
// Input:
// * mobile: string;
// * deviceToken?: string; - Required for Android app.
func (router AuthRouter) MobileSignUp(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params input.MobileSignUpParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Create account from input data.
	baseAccount := account.NewMobileBaseAccount(params)
	// Save it.
	err := router.userRepo.CreateAccount(baseAccount)
	if err != nil {
		sugar.Error(err)
		// Check for duplicate.
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(render.NewVEAlreadyExists("email"))
			return
		}
		_ = render.New(w).DBError(err)
		return
	}

	// Tracking.
	clientApp := footprint.NewClient(req)
	fp := footprint.New(baseAccount.FtcID, clientApp).
		FromSignUp().
		WithAuth(enum.LoginMethodMobile, params.DeviceToken)

	go func() {
		err := router.userRepo.SaveFootprint(fp)
		if err != nil {
			sugar.Error()
		}
	}()

	_ = render.New(w).OK(reader.Account{
		BaseAccount: baseAccount,
		LoginMethod: enum.LoginMethodMobile,
		Wechat:      account.Wechat{},
		Membership:  reader.Membership{},
	})
}
