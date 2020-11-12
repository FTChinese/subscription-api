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
	"github.com/stripe/stripe-go"
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

	stripe.Key = cfg.MustStripeAPIKey()

	myDB := db.MustNewMySQL(config.MustMySQLMasterConn(s.Production))
	rdb := db.NewRedis(config.MustRedisAddress().Pick(s.Production))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postoffice.New(config.MustGetHanqiConn())

	guard := access.NewGuard(myDB)

	payRouter := controller.NewSubsRouter(myDB, promoCache, cfg, post, logger)
	iapRouter := controller.NewIAPRouter(myDB, rdb, logger, post, cfg)
	stripeRouter := controller.NewStripeRouter(myDB, cfg, logger)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := controller.NewPaywallRouter(myDB, promoCache, logger)

	wxAuth := controller.NewWxAuth(myDB, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
	r.Use(controller.NoCache)

	// Requires user id.
	r.Route("/wxpay", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Use(controller.UserOrUnionID)

		// Create a new subscription for desktop browser
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/desktop/{tier}/{cycle}", payRouter.PlaceWxOrder(wechat.TradeTypeDesktop))
		r.With(controller.UserOrUnionID).
			Post("/desktop", payRouter.PlaceWxOrder(wechat.TradeTypeDesktop))

		// Create an order for mobile browser
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/mobile/{tier}/{cycle}", payRouter.PlaceWxOrder(wechat.TradeTypeMobile))
		r.With(controller.UserOrUnionID).
			Post("/mobile", payRouter.PlaceWxOrder(wechat.TradeTypeMobile))

		// Create an order for wx-embeded browser
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/jsapi/{tier}/{cycle}", payRouter.PlaceWxOrder(wechat.TradeTypeJSAPI))
		r.With(controller.UserOrUnionID).
			Post("/jsapi", payRouter.PlaceWxOrder(wechat.TradeTypeJSAPI))

		// Creat ean order for native app
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/app/{tier}/{cycle}", payRouter.PlaceWxOrder(wechat.TradeTypeApp))
		r.With(controller.UserOrUnionID).
			Post("/app", payRouter.PlaceWxOrder(wechat.TradeTypeApp))

		// Query order
		// Deprecated
		r.Get("/query/{id}", payRouter.VerifyPayment(false))
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Create an order for desktop browser
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/desktop/{tier}/{cycle}", payRouter.PlaceAliOrder(ali.EntryDesktopWeb))
		r.With(controller.UserOrUnionID).
			Post("/desktop", payRouter.PlaceAliOrder(ali.EntryDesktopWeb))

		// Create an order for mobile browser
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/mobile/{tier}/{cycle}", payRouter.PlaceAliOrder(ali.EntryMobileWeb))
		r.With(controller.UserOrUnionID).
			Post("/mobile", payRouter.PlaceAliOrder(ali.EntryMobileWeb))

		// Create an order for native app.
		// Deprecated
		r.With(controller.UserOrUnionID).
			Post("/app/{tier}/{cycle}", payRouter.PlaceAliOrder(ali.EntryApp))
		r.With(controller.UserOrUnionID).
			Post("/app", payRouter.PlaceAliOrder(ali.EntryApp))

		// Deprecated.
		r.Get("/query/{id}", payRouter.VerifyPayment(false))
	})

	r.Route("/upgrade", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.UserOrUnionID)
		r.Get("/balance", payRouter.PreviewUpgrade)
		// Get membership information when user want to upgrade: days remaining, account balance, amount
		r.Put("/free", payRouter.FreeUpgrade)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// List a user's orders
		//r.Get("/", payRouter.ListOrders)
		//r.Get("/{id}", payRouter.LoadOrder)

		// Get a payment's result from providers.
		r.Get("/{id}/payment-result", payRouter.VerifyPayment(true))
		// Verify if an order is confirmed, and returns PaymentResult.
		r.Post("/{id}/verify-payment", payRouter.VerifyPayment(false))

		// Manually confirm an order if not confirmed yet by verifying against
		// alipay of wxpay APIs. If it's already confirmed, nothing changes.
		// Deprecated.
		r.Patch("/{id}", payRouter.VerifyPayment(false))
	})

	r.Route("/stripe", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.FtcID)

		// Get a stripe plan.
		r.Get("/plans/{id}", stripeRouter.GetPlan)

		r.Route("/customers", func(r chi.Router) {
			// Create a stripe customer if not exists yet, or
			// just return the customer id if already exists.
			// Deprecated
			r.Put("/", stripeRouter.CreateCustomer)
			// Get stripe user's default payment method.
			r.Get("/{id}/default_payment_method", stripeRouter.GetDefaultPaymentMethod)
			// Set stripe user's default payment method.
			r.Post("/{id}/default_payment_method", stripeRouter.SetDefaultPaymentMethod)
			// Generate ephemeral key for client when it is
			// trying to modify customer data.
			r.Post("/{id}/ephemeral_keys", stripeRouter.IssueKey)
		})

		// Create Stripe subscription.
		r.Route("/subscriptions", func(r chi.Router) {
			r.Get("/", stripeRouter.GetSubscription)
			r.Post("/", stripeRouter.CreateSubscription)
			// Upgrade membership.
			r.Patch("/", stripeRouter.UpgradeSubscription)
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
		r.Get("/subs", iapRouter.ListSubs)
		// Load a single subscription.
		r.Get("/subs/{id}", iapRouter.LoadSubs)
		// Refresh an existing subscription of an original transaction id.
		r.Patch("/subs/{id}", iapRouter.RefreshSubs)

		// Load a receipt and its associated subscription. Internal only.
		r.Get("/receipt/{id}", iapRouter.LoadReceipt)
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

	r.Route("/paywall", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Data used to build a paywall.
		r.Get("/", paywallRouter.LoadPaywall)
		// List all active pricing plans.
		r.Get("/pricing", paywallRouter.LoadPricing)
		// Bust cache.
		r.Get("/__refresh", paywallRouter.BustCache)
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
