package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/access"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/repository/wxoauth"
	"github.com/stripe/stripe-go"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/controller"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
)

var (
	cfg     config.BuildConfig
	version string
	build   string
	commit  string
)

const (
	port = "8200"
)

func init() {
	var production bool
	var sandbox bool
	flag.BoolVar(&production, "production", false, "Connect to production MySQL database if present. Default to localhost.")
	flag.BoolVar(&sandbox, "sandbox", false, "Use sandbox for alipay and wxpay webhook url and stripe keys")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	cfg = config.NewBuildConfig(production, sandbox)

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	log.Printf("Production %t. Sandbox %t", production, sandbox)

	config.MustSetupViper()
}

func main() {

	stripe.Key = cfg.MustStripeAPIKey()

	myDB := db.MustNewDB(cfg.MustGetDBConn("mysql.master"))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postoffice.New(config.MustGetHanqiConn())

	guard := access.NewGuard(myDB)

	payRouter := controller.NewPayRouter(myDB, promoCache, cfg, post)
	upgradeRouter := controller.NewUpgradeRouter(payRouter)

	stripeRouter := controller.NewStripeRouter(myDB, cfg)
	iapRouter := controller.NewIAPRouter(myDB, cfg, post)

	//giftCardRouter := controller.NewGiftCardRouter(myDB, cfg)
	paywallRouter := controller.NewPaywallRouter(myDB, promoCache, cfg)

	wxAuth := controller.NewWxAuth(wxoauth.New(myDB))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
	r.Use(controller.NoCache)

	r.Get("/__version", status)

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
		// X-App-Id
		r.Get("/query/{orderId}", payRouter.QueryWxOrder)
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

		r.Get("/query/{orderId}", payRouter.QueryAliOrder)
	})

	r.Route("/upgrade", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.UserOrUnionID)
		// Get membership information when user want to upgrade: days remaining, account balance, amount
		r.Put("/free", upgradeRouter.FreeUpgrade)
		r.Get("/balance", upgradeRouter.UpgradeBalance)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(guard.CheckToken)

		// Manually confirm an order if not confirmed yet by verifying against
		// alipay of wxpay APIs. If it's already confirmed, nothing changes.
		r.Put("/:id", payRouter.ManualConfirm)
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

		// Update subscription based on the passed in receipt data.
		// The only difference between this one and /verify-receipt
		// is the response data.
		r.Post("/subscription", iapRouter.UpsertSubs)
		//r.Get("/subscription", iapRouter.LoadSubs)
		// Refresh an existing subscription of an original transaction id.
		r.Patch("/subscription/{id}", iapRouter.RefreshSubs)

		// Link FTC account to apple subscription.
		// This step does not perform verification.
		// It only links an existing subscription to ftc account.
		// You should ask the /subscription endpoint to
		// update data and get the original transaction id.
		r.Post("/link", iapRouter.Link)
		// Unlink ftc account from apple subscription.
		r.Delete("/link", iapRouter.Unlink)

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

			r.Get("/callback", wxAuth.WebCallback)
		})
	})

	log.Printf("Subscription api running at %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func status(w http.ResponseWriter, _ *http.Request) {

	data := struct {
		Version    string `json:"version"`
		Build      string `json:"build"`
		Commit     string `json:"commit"`
		Production bool   `json:"production"`
		Sandbox    bool   `json:"sandbox"`
	}{
		Version:    version,
		Build:      build,
		Commit:     commit,
		Production: cfg.Production(),
		Sandbox:    cfg.Sandbox(),
	}

	_ = render.New(w).OK(data)
}
