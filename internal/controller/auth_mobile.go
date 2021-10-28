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
	// In such case we will treat the mobile already exists
	// and the mobile should be synced between tables, which is
	// delayed till the account is fetched.
	if result.ID.Valid {
		go func() {
			fp := footprint.
				New(result.ID.String, footprint.NewClient(req)).
				FromLogin().
				WithAuth(enum.LoginMethodMobile, params.DeviceToken)

			err = router.userRepo.SaveFootprint(fp)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(result)
}

// MobileLinkExistingEmail authenticates an existing email account,
// and link to the provided mobile phone.
//
// Input:
// * email: string;
// * password: string;
// * mobile: string;
// * deviceToken?: string;
//
// Require header footprint.Client.
//
// Returns reader.Account.
// Possible cases:
// * The link target does not exist in profile table. Insert.
// * The link target present in profile table but mobile column missing. Update.
// * The link target present in profile table and mobile column is taken by another mobile number. Deny.
// * The link target present in profile and mobile column is this mobile number. No action.
func (router AuthRouter) MobileLinkExistingEmail(w http.ResponseWriter, req *http.Request) {
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
	authResult, err := router.userRepo.Authenticate(params.EmailCredentials)
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

	// Retrieve account by user id.
	// There shouldn't be any 404 error here.
	// We should make sure this account does not have any
	// mobile attached before linking.
	acnt, err := router.userRepo.AccountByFtcID(authResult.UserID)
	if err != nil {
		// There shouldn't be ErrNoRow error here.
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	currentMobile := acnt.GetMobile()
	if currentMobile != "" {
		// Mobile already set. Return the account immediately.
		if currentMobile == params.Mobile {
			_ = render.New(w).OK(acnt)
			return
		}

		// Account linked to another mobile.
		// We do not allow overriding existing email account's
		// mobile here.
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "This email account already has another mobile set",
			Field:   "mobile",
			Code:    render.CodeAlreadyExists,
		})
		return
	}

	// Up till now we can make sure this email account does
	// not have mobile set.
	// We have to make sure the mobile does not exist
	// in current db.
	mobileSearch, err := router.userRepo.SearchByMobile(params.Mobile)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// If found, the mobile must be linked to another account
	// since if it is linked to the requested email account,
	// it has already stopped in previous step.
	if mobileSearch.ID.Valid {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "The mobile already used by another account",
			Field:   "mobile",
			Code:    render.CodeAlreadyExists,
		})

		return
	}

	err = router.userRepo.UpsertMobile(account.MobileUpdater{
		FtcID:  acnt.FtcID,
		Mobile: null.StringFrom(params.Mobile),
	})
	if err != nil {
		// Ensure the new mobile is not used by anyone.
		if errors.Is(err, account.ErrMobileTakenByOther) {
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: err.Error(),
				Field:   "mobile",
				Code:    render.CodeAlreadyExists,
			})
		}
		sugar.Error(err)
	}

	acnt.Mobile = null.StringFrom(params.Mobile)

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

	// Kept for backward compatible as this was previously
	// used to create a new email account when mobile is
	// used to log in for the first time.
	if params.HasCredentials() {
		router.emailSignUp(w, input.EmailSignUpParams{
			EmailCredentials: params.EmailCredentials,
			Mobile:           null.StringFrom(params.Mobile),
			DeviceToken:      params.DeviceToken,
			SourceURL:        params.SourceURL,
		}, footprint.NewClient(req))
		return
	}

	// No email+password provided. Use phone number to generate
	// a fake email.
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
