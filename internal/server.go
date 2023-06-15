package internal

import (
	"log"
	"net/http"
	"time"

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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/patrickmn/go-cache"
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
	myDBs := db.MustNewMyDBs()
	gormDBs := db.MustNewMultiGormDBs(s.Production)
	rdb := db.NewRedis(config.MustRedisAddress().Pick(s.Production))

	// Set the cache default expiration and cleanup interval both to 2 hours.
	cacheStore := cache.New(2*time.Hour, 2*time.Hour)
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
	ftcPayRoutes := api.NewFtcPayRoutes(
		myDBs,
		cacheStore,
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

	stripeRoutes := api.NewStripeRoutes(
		myDBs,
		cacheStore,
		logger,
		s.LiveMode)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := api.NewPaywallRouter(
		myDBs,
		cacheStore,
		logger,
		s.LiveMode)

	legalRoutes := api.NewLegalRepo(
		myDBs,
		logger)

	cmsRouter := api.NewCMSRouter(myDBs, s.LiveMode, logger)

	appRouter := api.NewAndroidRouter(
		myDBs,
		cacheStore,
		logger)

	wxAuth := api.NewWxAuth(myDBs, logger)

	guard := access.NewGuard(gormDBs)

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
		r.Post("/desktop", ftcPayRoutes.WxPay(wechat.TradeTypeDesktop))

		// Create an order for mobile browser
		r.Post("/mobile", ftcPayRoutes.WxPay(wechat.TradeTypeMobile))

		// Create an order for wx-embedded browser
		r.Post("/jsapi", ftcPayRoutes.WxPay(wechat.TradeTypeJSAPI))

		// Creat an order for native app
		r.Post("/app", ftcPayRoutes.WxPay(wechat.TradeTypeApp))
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		r.Use(xhttp.FormParsed)

		// Create an order for desktop browser
		r.Post("/desktop", ftcPayRoutes.AliPay(ali.EntryDesktopWeb))

		// Create an order for mobile browser
		r.Post("/mobile", ftcPayRoutes.AliPay(ali.EntryMobileWeb))

		// Create an order for native app.
		r.Post("/app", ftcPayRoutes.AliPay(ali.EntryApp))
	})

	r.Route("/membership", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		// Get the membership of a user
		r.Get("/", accountRouter.LoadMembership)
		r.Post("/addons", ftcPayRoutes.ClaimAddOn)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)

		// Pagination: page=<int>&per_page=<int>
		r.With(xhttp.FormParsed).
			Get("/", ftcPayRoutes.ListOrders)
		r.Get("/{id}", ftcPayRoutes.LoadOrder)

		// Transfer order query data from ali or wx api as is.
		r.Get("/{id}/payment-result", ftcPayRoutes.RawPaymentResult)
		// Verify if an order is confirmed, and returns PaymentResult.
		r.Post("/{id}/verify-payment", ftcPayRoutes.VerifyPayment)
	})

	r.Route("/ftc-pay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireFtcOrUnionID)
		r.Route("/invoices", func(r chi.Router) {
			// List a user's invoices. Use query parameter `kind=create|renew|upgrade|addon` to filter.
			r.Get("/", ftcPayRoutes.ListInvoices)
			// Show a single invoice.
			r.Get("/{id}", ftcPayRoutes.LoadInvoice)
		})
		r.Route("/discounts", func(r chi.Router) {
			r.Get("/{id}", ftcPayRoutes.LoadDiscountRedeemed)
		})
	})

	// All the following endpoints require `X-User-Id` header set except publishable-key and prices section.
	r.Route("/stripe", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Get("/publishable-key", stripeRoutes.PublishableKey)

		r.Route("/prices", func(r chi.Router) {
			r.Use(xhttp.FormParsed)
			// List stripe prices. If query parameter has refresh=true, no cached data will be used.
			// ?refresh=true|false
			r.Get("/", stripeRoutes.ListPaywallPrices)
			// Load a stripe price. It first queries ftc's db.
			// If not found, then query Stripe API.
			// Any price loade directly from Stripe API will
			// be inserted/updated in ftc's db.
			// Use query parameter `?refresh=true` to hit Stripe API directly.
			r.Get("/{id}", stripeRoutes.LoadPrice)
			// ?active_only=<true|false>
			// To create/update/delete a coupon, use the /cms section.`
			r.Get("/{id}/coupons", stripeRoutes.ListPriceCoupons)
		})

		r.Route("/coupons", func(r chi.Router) {
			r.Use(xhttp.FormParsed)
			// ?refresh=true
			r.Get("/{id}", stripeRoutes.LoadStripeCoupon)
		})

		r.Route("/customers", func(r chi.Router) {

			r.Use(xhttp.RequireFtcID)

			// Create a stripe customer if not exists yet
			r.Post("/", stripeRoutes.CreateCustomer)
			// Use this to check customer's default source and default payment method.
			// refresh=true
			r.Get("/{id}", stripeRoutes.GetCustomer)

			r.Post("/{id}/default-payment-method", stripeRoutes.UpdateCusDefaultPaymentMethod)
			// ?refresh=true
			r.Get("/{id}/default-payment-method", stripeRoutes.GetCusDefaultPaymentMethod)

			r.Get("/{id}/payment-methods", stripeRoutes.ListCusPaymentMethods)

			// Generate ephemeral key for client when it is
			// trying to modify customer data.
			r.Post("/{id}/ephemeral-keys", stripeRoutes.IssueKey)
		})

		r.Route("/setup-intents", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)
			// Create a payment method
			r.Post("/", stripeRoutes.CreateSetupIntent)
			// ?refresh=true
			r.With(xhttp.FormParsed).
				Get("/{id}", stripeRoutes.GetSetupIntent)
			// ?refresh=true
			r.Get("/{id}/payment-method", stripeRoutes.GetSetupPaymentMethod)
		})

		r.Route("/payment-sheet", func(r chi.Router) {
			r.Post("/setup", stripeRoutes.SetupWithEphemeral)
		})

		r.Route("/payment-methods", func(r chi.Router) {
			// Query parameter: ?refresh=true|false
			r.With(xhttp.FormParsed).
				Get("/{id}", stripeRoutes.LoadPaymentMethod)
		})

		r.Route("/subs", func(r chi.Router) {
			r.Use(xhttp.RequireFtcID)

			// Create a subscription
			r.Post("/", stripeRoutes.CreateSubs)
			// Get a single subscription
			r.Get("/{id}", stripeRoutes.LoadSubs)
			// Update a subscription
			r.Post("/{id}", stripeRoutes.UpdateSubs)
			r.Post("/{id}/refresh", stripeRoutes.RefreshSubs)
			r.Post("/{id}/cancel", stripeRoutes.CancelSubs)
			r.Post("/{id}/reactivate", stripeRoutes.ReactivateSubscription)
			r.Get("/{id}/default-payment-method", stripeRoutes.GetSubsDefaultPaymentMethod)
			r.Post("/{id}/default-payment-method", stripeRoutes.UpdateSubsDefaultPayMethod)
			r.Get("/{id}/latest-invoice", stripeRoutes.LoadLatestInvoice)
			r.Get("/{id}/latest-invoice/any-coupon", stripeRoutes.CouponOfLatestInvoice)
		})

		r.Route("/invoices", func(r chi.Router) {
			// ?refresh=true
			r.Get("/{id}", stripeRoutes.LoadInvoice)
		})
	})

	r.Route("/paywall", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Data used to build a paywall.
		// Live server only outputs live data while sandbox for sandbox data only.
		// ?refresh=true
		r.With(xhttp.FormParsed).Get("/", paywallRouter.LoadPaywall)
		r.Post("/__migrate/active_prices", paywallRouter.MigrateToActivePrices)

		// List active prices used on paywall.
		r.Get("/active/prices", paywallRouter.LoadFtcActivePrices)

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
			r.Get("/{id}", paywallRouter.LoadPrice)
			// Activate a price under a product. All its sibling price of same tier and kind will be deactivated.
			// To activate an introductory price ,use PATCH /products/{id}/intro.
			r.Post("/{id}/activate", paywallRouter.ActivatePrice)
			r.Post("/{id}/deactivate", paywallRouter.DeactivateOrArchivePrice(false))
			r.Patch("/{id}", paywallRouter.UpdatePrice)
			r.Patch("/{id}/discounts", paywallRouter.RefreshPriceOffers)
			r.Delete("/{id}", paywallRouter.DeactivateOrArchivePrice(true))
		})

		// List discounts of a price.
		// ?price_id=<string>
		r.Route("/discounts", func(r chi.Router) {
			// List discounts of specified price.
			// Query parameter: ?price_id=<price id>
			r.Get("/", paywallRouter.ListDiscounts)
			r.Get("/{id}", paywallRouter.LoadDiscount)
			// Creates a new discounts for a price.
			r.Post("/", paywallRouter.CreateDiscount)
			// Delete discount and refresh the related price.
			r.Delete("/{id}", paywallRouter.DropDiscount)
		})
	})

	r.Route("/webhook", func(r chi.Router) {
		r.Post("/wxpay", ftcPayRoutes.WxWebHook)
		r.Post("/alipay", ftcPayRoutes.AliWebHook)
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
		r.Post("/stripe", stripeRoutes.WebHook)
		r.Post("/apple", iapRouter.WebHook)
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

	r.Route("/apps", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Route("/android", func(r chi.Router) {
			// Use ?refresh=true to bust cache.
			r.Get("/latest", appRouter.LatestRelease)
			r.Get("/releases", appRouter.ListReleases)
			r.Get("/releases/{versionName}", appRouter.LoadOneRelease)
		})
	})

	r.Route("/legal", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Get("/", legalRoutes.ListActive)
		r.Get("/{id}", legalRoutes.Load)
	})

	// Isolate dangerous operations from user-facing features.
	r.Route("/cms", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(xhttp.RequireStaffName)

		r.Route("/orders", func(r chi.Router) {
			r.With(xhttp.FormParsed).
				Get("/", ftcPayRoutes.CMSListOrders)
			r.Get("/{id}", ftcPayRoutes.CMSListOrders)
		})

		r.Route("/memberships", func(r chi.Router) {
			// Create or update a membership for a user
			r.Post("/", cmsRouter.UpsertMembership)
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

		r.Route("/stripe", func(r chi.Router) {

			r.Route("/prices", func(r chi.Router) {
				// ?page=<int>&per_page=<int>
				r.With(xhttp.FormParsed).Get("/", stripeRoutes.ListPricesPaged)
				// Add some essential metadata to a stripe price.
				r.Patch("/{id}", stripeRoutes.SetPriceMeta)
				r.Patch("/{id}/activate", stripeRoutes.ActivatePrice)
				r.Patch("/{id}/deactivate", stripeRoutes.DeactivatePrice)
			})

			r.Route("/coupons", func(r chi.Router) {
				// Link a coupon to a price, or modify its metadata
				r.Post("/{id}", stripeRoutes.UpdateStripeCoupon)
				r.Patch("/{id}/activate", stripeRoutes.ActivateCoupon)
				r.Patch("/{id}/pause", stripeRoutes.PauseCoupon)
				// Delete coupon does not perform DB deletion operation.
				// It simply flags the status field to cancelled status.
				r.Delete("/{id}", stripeRoutes.DeleteCoupon)
			})
		})

		r.Route("/legal", func(r chi.Router) {
			r.Get("/", legalRoutes.ListAll)
			r.Post("/", legalRoutes.Create)
			r.Patch("/{id}", legalRoutes.Update)
			r.Post("/{id}/publish", legalRoutes.Publish)
		})

		r.Route("/android", func(r chi.Router) {
			r.Post("/", appRouter.CreateRelease)
			r.Patch("/{versionName}", appRouter.UpdateRelease)
			r.Delete("/{versionName}", appRouter.DeleteRelease)
		})
	})

	r.Get("/__version", func(w http.ResponseWriter, req *http.Request) {
		_ = render.New(w).OK(s)
	})

	log.Fatal(http.ListenAndServe(":"+s.Port, r))
}
