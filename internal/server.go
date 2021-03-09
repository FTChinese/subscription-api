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

	rwMyDB := db.MustNewMySQL(config.MustMySQLMasterConn(s.Production))
	// DB connection with delete privilege.
	rwdMyDB := db.MustNewMySQL(config.MustMySQLAPIConn(s.Production))

	rdb := db.NewRedis(config.MustRedisAddress().Pick(s.Production))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postoffice.New(config.MustGetHanqiConn())

	guard := access.NewGuard(rwMyDB)

	payRouter := controller.NewSubsRouter(rwdMyDB, promoCache, cfg, post, logger)
	iapRouter := controller.NewIAPRouter(rwdMyDB, rdb, logger, post, cfg)
	stripeRouter := controller.NewStripeRouter(rwMyDB, cfg, logger)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := controller.NewPaywallRouter(rwMyDB, promoCache, logger)

	wxAuth := controller.NewWxAuth(rwMyDB, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
	r.Use(controller.NoCache)

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

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// List a user's orders
		//r.Get("/", payRouter.ListOrders)
		//r.Get("/{id}", payRouter.LoadOrder)

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

	// Deprecate. Use /webhook
	r.Route("/callback", func(r1 chi.Router) {
		r1.Post("/wxpay", payRouter.WxWebHook)
		r1.Post("/alipay", payRouter.AliWebHook)
	})

	r.Route("/webhook", func(r chi.Router) {
		r.Post("/wxpay", payRouter.WxWebHook)
		r.Post("/alipay", payRouter.AliWebHook)
		r.Post("/stripe", stripeRouter.WebHook)
		r.Post("/apple", iapRouter.WebHook)
	})

	// Handle wechat oauth.
	r.Route("/wx", func(r chi.Router) {

		r.Route("/oauth", func(r chi.Router) {

			r.With(controller.RequireAppID).Post("/login", wxAuth.Login)

			r.With(controller.RequireAppID).Put("/refresh", wxAuth.Refresh)

			// Do not check access token here since it is used by wx.
			r.Get("/callback", wxAuth.WebCallback)
		})
	})

	r.Get("/__version", func(w http.ResponseWriter, req *http.Request) {
		_ = render.New(w).OK(s)
	})

	log.Fatal(http.ListenAndServe(":"+s.Port, r))
}
