package internal

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/access"
	"github.com/FTChinese/subscription-api/internal/controller"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"time"
)

type ServerStatus struct {
	Version    string `json:"version"`
	Build      string `json:"build"`
	Commit     string `json:"commit"`
	Port       string `json:"-"`
	Production bool   `json:"production"` // Determine which db to use.
	LiveMode   bool   `json:"liveMode"`
}

func StartServer(s ServerStatus) {
	logger := config.MustGetLogger(s.Production)

	myDBs := db.MustNewMyDBs(s.Production)

	rdb := db.NewRedis(config.MustRedisAddress().Pick(s.Production))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postman.New(config.MustGetHanqiConn())

	guard := access.NewGuard(myDBs)

	authRouter := controller.NewAuthRouter(
		myDBs,
		post,
		logger)
	accountRouter := controller.NewAccountRouter(
		myDBs,
		post,
		logger)
	payRouter := controller.NewSubsRouter(
		myDBs,
		promoCache,
		s.LiveMode,
		post,
		logger)
	iapRouter := controller.NewIAPRouter(
		myDBs,
		rdb,
		logger,
		post,
		s.LiveMode)
	stripeRouter := controller.NewStripeRouter(
		myDBs,
		logger,
		s.LiveMode)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := controller.NewPaywallRouter(
		myDBs,
		promoCache,
		logger)

	wxAuth := controller.NewWxAuth(myDBs, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.DumpRequest)
	r.Use(controller.NoCache)

	r.Route("/auth", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Route("/email", func(r chi.Router) {
			// Checks if an email exists.
			// The email parameter should be sent as a query parameter `?v=<email>`
			// Returns HTTP status code 204 if the email exists,
			// or 404 if not found.
			r.Get("/exists", authRouter.EmailExists)
			// Authenticate user's email + password combination.
			r.Post("/login", authRouter.EmailLogin)
			// Create a new account using the provided email + password
			// When user login with mobile for the 1st time,
			// choose to sign up with a new email, it is
			// also handled by this endpoint.
			r.Post("/signup", authRouter.EmailSignUp)
			// Verify user's email by checking the validity of
			// a token send to user's email.
			// This is used only in desktop browsers.
			r.Post("/verification/{token}", authRouter.VerifyEmail)
		})

		r.Route("/mobile", func(r chi.Router) {
			// Create a SMS code and send it to user for login.
			// This differ from /account/mobile/verification in that
			// there is not user id set in header; therefore the record
			// in DB does not have user id field save alongside the
			// code.
			// After getting mobile number, we first tries to retrieve
			// the minimal account data. If found, we will save the
			// code and user id together; otherwise we user id field
			// won't exist along with the code, which indicates the
			// user is logging in using mobile for the first  time.
			r.Put("/verification", authRouter.RequestSMSVerification)
			// Verifies an SMS code. If the code is found, a nullable
			// user id associated with the code is returned.
			// There are 3 choices to follow depending on the
			// user id nullability:
			// * If user id is not null, use the /account endpoint to retrieve data;
			// * If user is null, it indicates we didn't find the phone, client could:
			//   - Let user to signup with the phone, which will create a faked email derived from this phone number;
			//   - Let user link to an existing email account;
			//   - Let user sign up with a new email+password just as the email signup workflow,
			//     and this new email account will have mobile attached.
			r.Post("/verification", authRouter.VerifySMSCode)
			// Verifies an existing email account credentials.
			// If passed, retrieve user's full account and check if
			// the mobile is set to another one. If it is taken
			// by another one, returns 422; otherwise update
			// the account's mobile field and returns it.
			// In background thread we persist phone number to db.
			r.Post("/link", authRouter.MobileLinkExistingEmail)
			// When user login with mobile for the first time,
			// and has no account previously created, create a
			// new email account with mobile number set to the
			// specified one.
			// FIX: this actually should not be named as signup.
			// It combines two actions: signup with email and then perform
			// mobile to email link.
			// Signup should use mobile derived email to create a new account.
			r.Post("/signup", authRouter.MobileSignUp)
		})

		r.Route("/password-reset", func(r chi.Router) {
			r.Post("/", authRouter.ResetPassword)
			r.Post("/letter", authRouter.ForgotPassword)
			r.Get("/tokens/{token}", authRouter.VerifyResetToken)
			r.Get("/codes", authRouter.VerifyResetCode)
		})

		r.Route("/wx", func(r chi.Router) {
			r.Use(controller.RequireAppID)
			r.Post("/login", wxAuth.Login)
			r.Post("/refresh", wxAuth.Refresh)
		})
	})

	// A dedicated root path to handle oauth callback to avoid access checking.
	r.Route("/oauth", func(r chi.Router) {
		// Callback for web to get oauth code.
		// Do not check access token here since it is used by wx.
		// Deprecated.
		r.Route("/wx/callback", func(r chi.Router) {
			r.Get("/next-reader", controller.WxCallbackHandler(wxlogin.CallbackAppNextUser))
		})

		r.Route("/callback", func(r chi.Router) {
			r.Use(controller.FormParsed)
			r.Route("/wx", func(r chi.Router) {
				r.Get("/next-user", controller.WxCallbackHandler(wxlogin.CallbackAppNextUser))
				r.Get("/fta-reader", controller.WxCallbackHandler(wxlogin.CallbackAppFtaReader))
			})
		})
	})

	r.Route("/account", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Get account by uuid.
		r.With(controller.RequireFtcID).
			Get("/", accountRouter.LoadAccountByFtcID)

		r.With(controller.RequireFtcID).
			Delete("/", accountRouter.DeleteFtcAccount)

		r.Route("/email", func(r chi.Router) {
			r.Use(controller.RequireFtcID)

			// Update email.
			// Possible issues when user is also a Stripe customer:
			// the email we have at hand will be inconsistent with Stripe customer's email.
			// However this is a not a big problem since customer id is not affected.
			// We can still find this customer.
			r.Patch("/", accountRouter.UpdateEmail)
			r.Post("/request-verification", accountRouter.RequestVerification)
		})

		r.With(controller.RequireFtcID).
			Patch("/name", accountRouter.UpdateName)

		r.Route("/password", func(r chi.Router) {
			r.Use(controller.RequireFtcID)

			// Update password.
			r.Patch("/", accountRouter.UpdatePassword)
			//r.Post("/verification", accountRouter.VerifyPassword)
		})

		r.Route("/mobile", func(r chi.Router) {
			r.Use(controller.RequireFtcID)
			r.Post("/", accountRouter.DeleteMobile)
			// Set/Update mobile number by verifying SMS code.
			r.Patch("/", accountRouter.UpdateMobile)
			// Create a verification code for a logged-in user.
			// It differs from /auth/mobile/verification in that
			// this one requires user id being set in header.
			// When creating a record in DB, user id is saved alongside
			// the SMS code so that later when performing verification,
			// we could verify this code is indeed target at this user.
			r.Put("/verification", accountRouter.SMSToModifyMobile)
		})

		r.Route("/address", func(r chi.Router) {
			r.Use(controller.RequireFtcID)
			r.Get("/", accountRouter.LoadAddress)
			r.Patch("/", accountRouter.UpdateAddress)
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(controller.RequireFtcID)
			r.Get("/", accountRouter.LoadProfile)
			r.Patch("/", accountRouter.UpdateProfile)
		})

		r.Route("/wx", func(r chi.Router) {
			r.Use(controller.RequireUnionID)
			r.Get("/", accountRouter.LoadAccountByWx)
			r.Post("/signup", accountRouter.WxSignUp)
			// Wechat logged-in user links to an existing email account,
			// or email logged-in user links to wechat after authorization.
			// For wechat user, you must first verify email + password just like
			// user is logging in. Get the email account full data and inspect
			// if the links is permitted on the client side. Then sent link request here.
			r.Post("/link", accountRouter.WxLinkEmail)
			r.Post("/unlink", accountRouter.WxUnlinkEmail)
		})
	})

	// Requires user id.
	r.Route("/wxpay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)

		// Create a new subscription for desktop browser
		r.Post("/desktop", payRouter.WxPay(wechat.TradeTypeDesktop))

		// Create an order for mobile browser
		r.Post("/mobile", payRouter.WxPay(wechat.TradeTypeMobile))

		// Create an order for wx-embedded browser
		r.Post("/jsapi", payRouter.WxPay(wechat.TradeTypeJSAPI))

		// Creat an order for native app
		r.Post("/app", payRouter.WxPay(wechat.TradeTypeApp))
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)
		r.Use(controller.FormParsed)

		// Create an order for desktop browser
		r.Post("/desktop", payRouter.AliPay(ali.EntryDesktopWeb))

		// Create an order for mobile browser
		r.Post("/mobile", payRouter.AliPay(ali.EntryMobileWeb))

		// Create an order for native app.
		r.Post("/app", payRouter.AliPay(ali.EntryApp))
	})

	r.Route("/membership", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)
		// Get the membership of a user
		r.Get("/", payRouter.LoadMembership)
		// Update the membership of a user
		r.Patch("/", payRouter.UpdateMembership)
		// Create a membership of a user
		r.Put("/", payRouter.CreateMembership)
		// List the modification history of a user's membership
		r.Get("/snapshots", payRouter.ListMemberSnapshots)
		r.Post("/addons", payRouter.ClaimAddOn)
		r.Patch("/addons", payRouter.CreateAddOn)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)

		// Pagination: page=<int>&per_page=<int>
		r.Get("/", payRouter.ListOrders)
		r.Get("/{id}", payRouter.LoadOrder)

		// Transfer order query data from ali or wx api as is.
		r.Get("/{id}/payment-result", payRouter.RawPaymentResult)
		// Verify if an order is confirmed, and returns PaymentResult.
		r.Post("/{id}/verify-payment", payRouter.VerifyPayment)
	})

	r.Route("/invoices", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)
		// List a user's invoices. Use query parameter `kind=create|renew|upgrade|addon` to filter.
		r.Get("/", payRouter.ListInvoices)
		r.Put("/", payRouter.CreateInvoice)
		// Show a single invoice.
		r.Get("/{id}", payRouter.LoadInvoice)
	})

	r.Route("/stripe", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Route("/prices", func(r chi.Router) {
			r.Use(controller.FormParsed)
			// List stripe prices. If query parameter has refresh=true, no cached data will be used.
			// ?refresh=true|false
			r.Get("/", stripeRouter.ListPrices)
			r.Get("/{id}", stripeRouter.LoadPrice)
		})

		r.Route("/customers", func(r chi.Router) {

			r.Use(controller.RequireFtcID)

			// Create a stripe customer if not exists yet, or
			// just return the customer id if already exists.
			r.Post("/", stripeRouter.CreateCustomer)
			// TODO: Set an existing customer id to the user if not set yet.
			// Use this to check customer's default source and default payment method.
			r.Get("/{id}", stripeRouter.GetCustomer)
			r.Post("/{id}/default-payment-method", stripeRouter.ChangeDefaultPaymentMethod)
			// Generate ephemeral key for client when it is
			// trying to modify customer data.
			r.Post("/{id}/ephemeral-keys", stripeRouter.IssueKey)
		})

		r.Route("/setup-intents", func(r chi.Router) {
			r.Use(controller.RequireFtcID)
			r.Post("/", stripeRouter.CreateSetupIntent)
		})

		r.Route("/checkout", func(r chi.Router) {
			r.Use(controller.RequireFtcID)

			r.Post("/", stripeRouter.CreateCheckoutSession)
		})

		r.Route("/subs", func(r chi.Router) {
			r.Use(controller.RequireFtcID)

			// Create a subscription
			r.Post("/", stripeRouter.CreateSubs)
			// List all subscriptions of a user
			r.Get("/", stripeRouter.ListSubs)
			// Get a single subscription
			r.Get("/{id}", stripeRouter.LoadSubs)
			r.Post("/{id}", stripeRouter.UpdateSubs)
			// Update a subscription
			r.Post("/{id}/refresh", stripeRouter.RefreshSubs)
			r.Post("/{id}/cancel", stripeRouter.CancelSubs)
			r.Post("/{id}/reactivate", stripeRouter.ReactivateSubscription)
		})
	})

	r.Route("/apple", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Verify an encoded receipt and returns the decoded data.
		r.Post("/verify-receipt", iapRouter.VerifyReceipt)
		// Link FTC account to apple subscription.
		r.Post("/link", iapRouter.Link)
		// Unlink ftc account from apple subscription.
		r.Post("/unlink", iapRouter.Unlink)

		// Verify a receipt like the verify-receipt.
		// Returns the extracted subscription instead the verified receipt.
		r.Post("/subs", iapRouter.UpsertSubs)
		// Load one subscription.

		// ?page=<int>&per_page<int>
		r.With(controller.RequireFtcID).Get("/subs", iapRouter.ListSubs)
		// Load a single subscription.
		r.Get("/subs/{id}", iapRouter.LoadSubs)
		// Refresh an existing subscription of an original transaction id.
		r.Patch("/subs/{id}", iapRouter.RefreshSubs)

		// Load a receipt and its associated subscription. Internal only.
		// ?fs=true
		r.Get("/receipt/{id}", iapRouter.LoadReceipt)
	})

	r.Route("/paywall", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Data used to build a paywall.
		// ?live=<true|false> to get prices for different mode.
		// TODO: in v5 this behavior will be dropped.
		// Live server only outputs live data while sandbox for sandbox data only.
		r.With(controller.FormParsed).
			Get("/", paywallRouter.LoadPaywall)

		// List active prices used on paywall.
		// ?live=<true|false>
		r.With(controller.FormParsed).
			Get("/active/prices", paywallRouter.LoadPricing)

		// The following are used by CMS to create/update prices and discounts.
		// Get a list of prices under a product. This does not distinguish is_active or live_mode
		// ?product_id=<string>
		r.With(controller.FormParsed).
			Get("/prices", paywallRouter.ListPrices)
		// Create a price for a product. The price's live mode is determined by client.
		r.Post("/prices", paywallRouter.CreatePrice)

		r.Post("/prices/{id}/activate", paywallRouter.ActivatePrice)
		r.Post("/prices/{id}/refresh", paywallRouter.RefreshPrice)

		// Retrieve all discounts of a price and save in under price row as JSON.
		// A price should only retrieve discount of the same live mode.
		// TODO: changed to update price.
		r.Patch("/prices/{id}", paywallRouter.RefreshPrice)
		r.Delete("/prices/{id}", paywallRouter.ArchivePrice)

		// List discounts of a price.
		// ?price_id=<string>
		r.Get("/discounts", paywallRouter.ListDiscounts)
		// Creates a new discounts for a price.
		r.Post("/discounts", paywallRouter.CreateDiscount)
		// Delete discount and refresh the related price.
		r.Delete("/discounts/{id}", paywallRouter.RemoveDiscount)

		// Bust cache, regardless of live mode or not.
		r.Get("/__refresh", paywallRouter.BustCache)
	})

	r.Route("/webhook", func(r chi.Router) {
		r.Post("/wxpay", payRouter.WxWebHook)
		r.Post("/alipay", payRouter.AliWebHook)
		// Events
		//invoice.finalized
		//invoice.payment_succeeded
		//invoice.payment_failed
		//invoice.created
		//customer.subscription.deleted
		//customer.subscription.updated
		//customer.subscription.created
		// http://www.ftacademy.cn/api/v1/webhook/stripe For version 1
		// http://www.ftacademy.cn/api/v2/webhook/stripe For version 2
		r.Post("/stripe", stripeRouter.WebHook)
		r.Post("/apple", iapRouter.WebHook)
	})

	r.Get("/__version", func(w http.ResponseWriter, req *http.Request) {
		_ = render.New(w).OK(s)
	})

	log.Fatal(http.ListenAndServe(":"+s.Port, r))
}
