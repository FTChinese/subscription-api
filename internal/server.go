package internal

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/access"
	"github.com/FTChinese/subscription-api/internal/app/api"
	"github.com/FTChinese/subscription-api/internal/pkg/letter"
	"github.com/FTChinese/subscription-api/internal/repository/accounts"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
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
	paywallCache := cache.New(2*time.Hour, 0)
	emailService := letter.NewService(logger)

	readerBaseRepo := shared.NewReaderCommon(myDBs)
	userShared := api.UserShared{
		Repo:         accounts.New(myDBs, logger),
		ReaderRepo:   readerBaseRepo,
		SMSClient:    ztsms.NewClient(logger),
		Logger:       logger,
		EmailService: emailService,
	}

	authRouter := api.NewAuthRouter(userShared)
	accountRouter := api.NewAccountRouter(userShared)
	ftcSubsRouter := api.NewFtcPayRouter(
		myDBs,
		paywallCache,
		logger,
		s.LiveMode)

	iapRouter := api.IAPRouter{
		Repo:         iaprepo.New(myDBs, rdb, logger),
		Client:       iaprepo.NewClient(logger),
		ReaderRepo:   readerBaseRepo,
		EmailService: emailService,
		Logger:       logger,
		Live:         s.LiveMode,
	}

	stripeRouter := api.NewStripeRouter(
		myDBs,
		paywallCache,
		logger,
		s.LiveMode)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := api.NewPaywallRouter(
		myDBs,
		paywallCache,
		logger,
		s.LiveMode)

	cmsRouter := api.NewCMSRouter(
		myDBs,
		paywallCache,
		logger,
		s.LiveMode)

	appRouter := api.NewAppRouter(myDBs)

	wxAuth := api.NewWxAuth(myDBs, logger)

	guard := access.NewGuard(myDBs)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(xhttp.DumpRequest)
	r.Use(xhttp.NoCache)

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
			// After getting mobile number, we first try to retrieve
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
			r.Use(xhttp.RequireAppID)
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
			r.Get("/next-reader", api.WxCallbackHandler(wxlogin.CallbackAppNextUser))
		})

		r.Route("/callback", func(r chi.Router) {
			r.Use(xhttp.FormParsed)
			r.Route("/wx", func(r chi.Router) {
				r.Get("/next-user", api.WxCallbackHandler(wxlogin.CallbackAppNextUser))
				r.Get("/fta-reader", api.WxCallbackHandler(wxlogin.CallbackAppFtaReader))
			})
		})
	})

	r.Route("/account", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Get account by uuid.
		r.With(xhttp.RequireFtcID).
			Get("/", accountRouter.LoadAccountByFtcID)

		r.With(xhttp.RequireFtcID).
			Delete("/", accountRouter.DeleteFtcAccount)

		r.Route("/email", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)

			// Update email.
			// Possible issues when user is also a Stripe customer:
			// the email we have at hand will be inconsistent with Stripe customer's email.
			// However this is a not a big problem since customer id is not affected.
			// We can still find this customer.
			r.Patch("/", accountRouter.UpdateEmail)
			r.Post("/request-verification", accountRouter.RequestVerification)
		})

		r.With(xhttp.RequireFtcID).
			Patch("/name", accountRouter.UpdateName)

		r.Route("/password", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)

			// Update password.
			r.Patch("/", accountRouter.UpdatePassword)
			//r.Post("/verification", accountRouter.VerifyPassword)
		})

		r.Route("/mobile", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)
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
			r.Use(xhttp.RequireFtcID)
			r.Get("/", accountRouter.LoadAddress)
			r.Patch("/", accountRouter.UpdateAddress)
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)
			r.Get("/", accountRouter.LoadProfile)
			r.Patch("/", accountRouter.UpdateProfile)
		})

		r.Route("/wx", func(r chi.Router) {
			r.Use(xhttp.RequireUnionID)
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
		r.Use(xhttp.RequireFtcOrUnionID)

		// Create a new subscription for desktop browser
		r.Post("/desktop", ftcSubsRouter.WxPay(wechat.TradeTypeDesktop))

		// Create an order for mobile browser
		r.Post("/mobile", ftcSubsRouter.WxPay(wechat.TradeTypeMobile))

		// Create an order for wx-embedded browser
		r.Post("/jsapi", ftcSubsRouter.WxPay(wechat.TradeTypeJSAPI))

		// Creat an order for native app
		r.Post("/app", ftcSubsRouter.WxPay(wechat.TradeTypeApp))
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		r.Use(xhttp.FormParsed)

		// Create an order for desktop browser
		r.Post("/desktop", ftcSubsRouter.AliPay(ali.EntryDesktopWeb))

		// Create an order for mobile browser
		r.Post("/mobile", ftcSubsRouter.AliPay(ali.EntryMobileWeb))

		// Create an order for native app.
		r.Post("/app", ftcSubsRouter.AliPay(ali.EntryApp))
	})

	r.Route("/membership", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		// Get the membership of a user
		r.Get("/", accountRouter.LoadMembership)
		r.Post("/addons", ftcSubsRouter.ClaimAddOn)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)

		// Pagination: page=<int>&per_page=<int>
		r.Get("/", ftcSubsRouter.ListOrders)
		r.Get("/{id}", ftcSubsRouter.LoadOrder)

		// Transfer order query data from ali or wx api as is.
		r.Get("/{id}/payment-result", ftcSubsRouter.RawPaymentResult)
		// Verify if an order is confirmed, and returns PaymentResult.
		r.Post("/{id}/verify-payment", ftcSubsRouter.VerifyPayment)
	})

	r.Route("/invoices", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		// List a user's invoices. Use query parameter `kind=create|renew|upgrade|addon` to filter.
		r.Get("/", ftcSubsRouter.ListInvoices)
		// Show a single invoice.
		r.Get("/{id}", ftcSubsRouter.LoadInvoice)
	})

	r.Route("/stripe", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Route("/prices", func(r chi.Router) {
			r.Use(xhttp.FormParsed)
			// List stripe prices. If query parameter has refresh=true, no cached data will be used.
			// ?refresh=true|false
			r.Get("/", stripeRouter.ListPrices)
			r.Get("/{id}", stripeRouter.LoadPrice)
		})

		r.Route("/customers", func(r chi.Router) {

			r.Use(xhttp.RequireFtcID)

			// Create a stripe customer if not exists yet
			r.Post("/", stripeRouter.CreateCustomer)
			// Use this to check customer's default source and default payment method.
			// refresh=true
			r.Get("/{id}", stripeRouter.GetCustomer)

			r.Post("/{id}/default-payment-method", stripeRouter.UpdateCusDefaultPaymentMethod)
			// ?refresh=true
			r.Get("/{id}/default-payment-method", stripeRouter.GetCusDefaultPaymentMethod)

			r.Get("/{id}/payment-methods", stripeRouter.ListCusPaymentMethods)

			// Generate ephemeral key for client when it is
			// trying to modify customer data.
			r.Post("/{id}/ephemeral-keys", stripeRouter.IssueKey)
		})

		r.Route("/setup-intents", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)
			// Create a payment method
			r.Post("/", stripeRouter.CreateSetupIntent)
			// ?refresh=true
			r.Get("/{id}", stripeRouter.GetSetupIntent)
			// ?refresh=true
			r.Get("/{id}/payment-method", stripeRouter.GetSetupPaymentMethod)
		})

		r.Route("/payment-sheet", func(r chi.Router) {
			r.Post("/setup", stripeRouter.SetupWithEphemeral)
		})

		r.Route("/payment-methods", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)
			// Query parameter: ?refresh=true|false
			r.Get("/{id}", stripeRouter.LoadPaymentMethod)
		})

		r.Route("/subs", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)

			// Create a subscription
			r.Post("/", stripeRouter.CreateSubs)
			// Get a single subscription
			r.Get("/{id}", stripeRouter.LoadSubs)
			r.Post("/{id}", stripeRouter.UpdateSubs)
			// Update a subscription
			r.Post("/{id}/refresh", stripeRouter.RefreshSubs)
			r.Post("/{id}/cancel", stripeRouter.CancelSubs)
			r.Post("/{id}/reactivate", stripeRouter.ReactivateSubscription)
			r.Get("/{id}/default-payment-method", stripeRouter.GetSubsDefaultPaymentMethod)
			r.Post("/{id}/default-payment-method", stripeRouter.UpdateSubsDefaultPayMethod)
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
		r.With(xhttp.RequireFtcID).Get("/subs", iapRouter.ListSubs)
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
		// Live server only outputs live data while sandbox for sandbox data only.
		r.Get("/", paywallRouter.LoadPaywall)

		// List active prices used on paywall.
		r.Get("/active/prices", paywallRouter.LoadPricing)

		r.Route("/banner", func(r chi.Router) {
			r.Post("/", paywallRouter.SaveBanner)
			r.Post("/promo", paywallRouter.SavePromo)
			r.Delete("/promo", paywallRouter.DropPromo)
		})

		r.Route("/products", func(r chi.Router) {
			r.Get("/", paywallRouter.ListProducts)
			r.Post("/", paywallRouter.CreateProduct)
			r.Get("/{id}", paywallRouter.LoadProduct)
			r.Patch("/{id}", paywallRouter.UpdateProduct)
			r.Post("/{id}/activate", paywallRouter.ActivateProduct)
			// Update the introductory price of a product.
			// The price will be set/overridden/refreshed on  the product
			// regardless of whether the product has one set or not.
			r.Patch("/{id}/intro", paywallRouter.AttachIntroPrice)
			// Delete intro price attached to a product.
			r.Delete("/{id}/intro", paywallRouter.DropIntroPrice)
		})

		// The following are used by CMS to create/update prices and discounts.
		r.Route("/prices", func(r chi.Router) {
			// Get a list of prices under a product. This does not distinguish is_active or live_mode
			// ?product_id=<string>
			r.With(xhttp.FormParsed).
				Get("/", paywallRouter.ListPrices)
			// Create a price for a product. The price's live mode is determined by client.
			r.Post("/", paywallRouter.CreatePrice)
			// Activate a price under a product. All its sibling price of same tier and kind will be deactivated.
			// To activate an introductory price ,use PATCH /products/{id}/intro.
			r.Post("/{id}/activate", paywallRouter.ActivatePrice)
			r.Patch("/{id}", paywallRouter.UpdatePrice)
			r.Patch("/{id}/discounts", paywallRouter.RefreshPriceOffers)
			r.Delete("/{id}", paywallRouter.ArchivePrice)
		})

		// List discounts of a price.
		// ?price_id=<string>
		r.Route("/discounts", func(r chi.Router) {
			r.Get("/", paywallRouter.ListDiscounts)
			// Creates a new discounts for a price.
			r.Post("/", paywallRouter.CreateDiscount)
			// Delete discount and refresh the related price.
			r.Delete("/{id}", paywallRouter.DropDiscount)
		})

		// Bust cache, regardless of live mode or not.
		r.Get("/__refresh", paywallRouter.BustCache)
	})

	r.Route("/apps", func(r chi.Router) {
		r.Route("/android", func(r chi.Router) {
			r.Get("/latest", appRouter.AndroidLatest)
			r.Get("/releases", appRouter.AndroidList)
			r.Get("/releases/{versionName}", appRouter.AndroidSingle)
		})
	})

	// Isolate dangerous operations from user-facing features.
	r.Route("/cms", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireStaffName)

		r.Route("/memberships", func(r chi.Router) {
			// Create a membership for a user
			r.Post("/", cmsRouter.CreateMembership)
			// Update the membership of a user
			r.Patch("/{id}", cmsRouter.UpdateMembership)
			r.Delete("/{id}", cmsRouter.DeleteMembership)
		})

		// ?ftc_id=<uuid>&union_id=<union_id>&page=<int>&per_page=<int>
		//r.With(xhttp.FormParsed).
		//	With(xhttp.RequireUserIDsQuery).
		//	Get("/snapshots", cmsRouter.ListMemberSnapshots)

		r.Route("/addons", func(r chi.Router) {
			// Add an invoice to a user.
			// If the invoice is targeting addon, then
			// membership should be updated accordingly.
			r.Post("/", cmsRouter.CreateAddOn)
		})
	})

	r.Route("/webhook", func(r chi.Router) {
		r.Post("/wxpay", ftcSubsRouter.WxWebHook)
		r.Post("/alipay", ftcSubsRouter.AliWebHook)
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
