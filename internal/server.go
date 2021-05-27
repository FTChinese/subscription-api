package internal

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/access"
	"github.com/FTChinese/subscription-api/internal/controller"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
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
	Production bool   `json:"production"`
	Sandbox    bool   `json:"sandbox"`
}

func StartServer(s ServerStatus) {
	cfg := config.NewBuildConfig(s.Production, s.Sandbox)
	logger := config.MustGetLogger(s.Production)

	myDBs := db.NewMyDB(s.Production)

	rdb := db.NewRedis(config.MustRedisAddress().Pick(s.Production))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postoffice.New(config.MustGetHanqiConn())

	guard := access.NewGuard(myDBs)

	authRouter := controller.NewAuthRouter(myDBs, post, logger)
	accountRouter := controller.NewAccountRouter(myDBs, post, logger)
	payRouter := controller.NewSubsRouter(myDBs, promoCache, cfg, post, logger)
	iapRouter := controller.NewIAPRouter(myDBs, rdb, logger, post, cfg)
	stripeRouter := controller.NewStripeRouter(myDBs, cfg, logger)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := controller.NewPaywallRouter(myDBs, promoCache, logger)

	wxAuth := controller.NewWxAuth(myDBs, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
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
			// Verifies a SMS code. If the code is found, a nullable
			// user id associated with the code is returned.
			// If the user id is null, it indicates the user is
			// logging in with mobile for the first time.
			// Client should then ask user to link to an existing
			// account, or create a new account accordingly.
			// If user id is not null, client should use the user to
			// retrieve account data from /account, providing the
			// user id in header.
			r.Post("/verification", authRouter.VerifySMSCode)
			// Verifies an existing email account credentials.
			// If passed, retrieve user's full account and check if
			// the mobile is set to another one. If it is taken
			// by another one, returns 422; otherwise update the
			// the account's mobile field and returns it.
			// In background thread we persist phone number to db.
			r.Post("/link", authRouter.LinkMobile)
			// When user login with mobile for the first time,
			// and has not account previously created, create a
			// new email account with mobile number set to the
			// specified one.
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
			r.Put("/refresh", wxAuth.Refresh)
		})
	})

	// Handle wechat oauth.
	// Deprecated. Use /auth/wx.
	r.Route("/wx", func(r chi.Router) {

		r.Route("/oauth", func(r chi.Router) {

			r.With(controller.RequireAppID).Post("/login", wxAuth.Login)

			r.With(controller.RequireAppID).Put("/refresh", wxAuth.Refresh)

			// Do not check access token here since it is used by wx.
			r.Get("/callback", wxAuth.WebCallback)
		})
	})

	r.Route("/oauth", func(r chi.Router) {
		// Callback for web to get oauth code.
		// Do not check access token here since it is used by wx.
		r.Route("/wx/callback", func(r chi.Router) {
			r.Get("/next-reader", wxAuth.WebCallback)
		})
	})

	r.Route("/account", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Get account by uuid.
		r.With(controller.RequireFtcID).Get("/", accountRouter.LoadAccountByEmail)

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

		r.With(controller.RequireFtcID).
			Patch("/password", accountRouter.UpdatePassword)

		r.Route("/mobile", func(r chi.Router) {
			r.Use(controller.RequireFtcID)
			// Set/Update mobile number by verifying SMS code.
			r.Patch("/", accountRouter.UpdateMobile)
			// Create a verification code for a logged-in user.
			// It differs from /auth/mobile/verification in that
			// this one requires user id being set in header.
			// When creating a record in DB, user id is save alongside
			// the SMS code so that later when performing verification,
			// we could verify this code is indeed target at this user.
			r.Put("/verification", accountRouter.RequestSMSVerification)
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
			r.Post("/link", accountRouter.LinkWechat)
			r.Post("/unlink", accountRouter.UnlinkWx)
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

	r.Route("/addon", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.RequireFtcOrUnionID)
		// List a list of add-on belonging to a user.
		//r.Get("/", )
		// Redeem add-on
		r.Post("/", payRouter.ClaimAddOn)
	})

	r.Route("/stripe", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// ?refresh=true|false
		r.Get("/prices", stripeRouter.ListPrices)

		r.With(controller.RequireFtcID).Route("/customers", func(r chi.Router) {

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

		r.With(controller.RequireFtcID).Route("/setup-intents", func(r chi.Router) {
			r.Post("/", stripeRouter.CreateSetupIntent)
		})

		r.With(controller.RequireFtcID).Route("/checkout", func(r chi.Router) {
			r.Post("/", stripeRouter.CreateCheckoutSession)
		})

		r.With(controller.RequireFtcID).Route("/subs", func(r chi.Router) {
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
		r.Get("/", paywallRouter.LoadPaywall)
		// List all active pricing plans.
		r.Get("/plans", paywallRouter.LoadPricing)
		// Bust cache.
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
