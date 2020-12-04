package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	"net/http"
)

// CreateSubs create a stripe subscription.
// Input:
// tier: string;
// cycle: string;
// coupon?: string;
// defaultPaymentMethod?: string;
// idempotency: string;
// Why this field?
//
// PITFALLS:
// If you creates a plan in CNY, and a customer is subscribed to
// it, and after that you created another plan in GBP, then
// Stripe will decline your subsequent subscription request.
// It's better to create different plans in the same currency.
// I guess Stripe takes plans in different currencies as the
// same one to avoid customer subscribing to the same plan
// in different countries and regions.
// {
// "status":400,
// "message":"You cannot combine currencies on a single customer. This customer has had a subscription, coupon, or invoice item with currency cny",
// "request_id":"req_fa0rfmytgnI22E",
// "type":"invalid_request_error"
// }
// in case user already linked wechat.
// Notification email is sent upon webhook receiving data, not here.
func (router StripeRouter) CreateSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(userIDKey)

	account, err := router.readerRepo.AccountByFtcID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound()
		return
	}

	var input ftcStripe.SubsInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validate input data.
	if err := input.Validate(false); err != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(err)
		return
	}

	sp, err := ftcStripe.PlanStore.FindByEditionV2(input.Edition, router.config.Live())
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Create stripe subscription.
	result, err := router.stripeRepo.CreateSubscription(ftcStripe.Checkout{
		Account: account,
		Plan:    sp,
		Params:  input.SubsParams,
	})

	if err != nil {
		sugar.Error(err)
		respErr, ok := reader.ParseStripeSubError(err)
		if ok {
			_ = render.New(w).JSON(respErr.StatusCode, respErr)
			return
		}

		err := forwardStripeErr(w, err)
		if err == nil {
			go func() {
				apiErr := ftcStripe.NewAPIError(input.FtcID, castStripeError(err))
				_ = router.stripeRepo.SaveSubsError(apiErr)
			}()

			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Save ftc id to stripe subscription id mapping.
	// Backup previous membership if exists.
	go func() {
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
	}()

	_ = render.New(w).OK(result)
}

// GetSubscription fetches a user's subscription and update membership if data in our db is stale.
//
// Error Response:
// 404: membership for this user is not found.
// Deprecated.
func (router StripeRouter) GetSubscription(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)

	mmb, err := router.readerRepo.RetrieveMember(reader.MemberID{
		FtcID: null.StringFrom(ftcID),
	}.MustNormalize())

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !mmb.IsStripe() {
		sugar.Error("Not a stripe membership")
		_ = render.New(w).NotFound()
	}

	subs, err := router.stripeRepo.RetrieveSubs(mmb.StripeSubsID.String)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(ftcStripe.NewSubs(subs))
}

func (router StripeRouter) ListSubs(w http.ResponseWriter, req *http.Request) {
	// TODO: implementation
}

func (router StripeRouter) LoadSubs(w http.ResponseWriter, req *http.Request) {

	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s, err := router.stripeRepo.RetrieveSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(s)
}

func (router StripeRouter) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s, err := router.stripeRepo.RetrieveSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	result, err := router.stripeRepo.RefreshSubs(s)
	if err != nil {
		return
	}

	_ = render.New(w).OK(result)
}

func (router StripeRouter) CancelSubs(w http.ResponseWriter, req *http.Request) {
	// TODO: implementation
}

func (router StripeRouter) UndoSubsCancel(w http.ResponseWriter, req *http.Request) {
	// TODO: implementation
}

// UpgradeSubscription create a stripe subscription.
// Input:
// tier: string;
// cycle: string;
// coupon?: "",
// defaultPaymentMethod?: "",
//
// Error response:
// 404 if membership if not found.
// NOTE: when updating a stripe subscription, the return payload
// `items` field contains more than one items:
// one is standard and another if premium.
// So we cannot rely on this field to find FTC plan.
func (router StripeRouter) UpgradeSubscription(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(userIDKey)

	account, err := router.readerRepo.AccountByFtcID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound()
		return
	}

	var input ftcStripe.SubsInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	if ve := input.Validate(true); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	sp, err := ftcStripe.PlanStore.FindByEditionV2(input.Edition, router.config.Live())
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := router.stripeRepo.UpgradeSubscription(ftcStripe.Checkout{
		Account: account,
		Plan:    sp,
		Params:  input.SubsParams,
	})

	if err != nil {
		err := forwardStripeErr(w, err)
		if err == nil {
			go func() {
				apiErr := ftcStripe.NewAPIError(input.FtcID, castStripeError(err))
				_ = router.stripeRepo.SaveSubsError(apiErr)
			}()

			return
		}

		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	go func() {
		err := router.stripeRepo.UpsertSubs(result.Subs)
		if err != nil {
			sugar.Error(err)
		}

		err = router.readerRepo.ArchiveMember(result.Snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(result)
}
