package main

import (
	"flag"
	"fmt"
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
	"github.com/FTChinese/go-rest/view"
	"github.com/FTChinese/subscription-api/controller"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	flag.BoolVar(&sandbox, "sandbox", false, "Use sandbox database to save subscription data if present.")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	cfg = config.NewBuildConfig(production, sandbox)

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.WithFields(logrus.Fields{
		"live":         cfg.Live(),
		"isProduction": cfg.IsProduction(),
	}).Infof("Initializing environment")

	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(1)
	}

	if cfg.Live() {
		stripe.Key = viper.GetString("stripe.live_secret_key")
	} else {
		stripe.Key = viper.GetString("stripe.test_secret_key")
	}
}

func main() {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "main",
	})

	stripe.Key = cfg.MustStripeAPIKey()

	myDB := db.MustNewDB(cfg.MustGetDBConn("mysql.master"))

	// Set the cache default expiration time to 2 hours.
	promoCache := cache.New(2*time.Hour, 0)

	post := postoffice.New(config.MustGetHanqiConn())

	guard := access.NewGuard(myDB)

	baseRouter := controller.NewBasePayRouter(myDB, promoCache, cfg, post)

	wxRouter := controller.NewWxRouter(baseRouter)
	aliRouter := controller.NewAliRouter(baseRouter)
	upgradeRouter := controller.NewUpgradeRouter(baseRouter)

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

		// Create a new subscription.
		r.Post("/desktop/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeDesktop))

		//r.Post("/desktop", wxRouter.NewSub(wechat.TradeTypeDesktop))
		//r.Patch("/desktop", wxRouter.UpgradeSub(wechat.TradeTypeDesktop))

		r.Post("/mobile/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeMobile))

		// {code: string}
		r.Post("/jsapi/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeJSAPI))

		r.Post("/app/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeApp))

		// Query order
		// X-App-Id
		r.Get("/query/{orderId}", wxRouter.OrderQuery)
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.UserOrUnionID)

		r.Post("/desktop/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryDesktopWeb))

		r.Post("/mobile/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryMobileWeb))

		r.Post("/app/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryApp))

		// Deprecate
		//r.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		// r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
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

		r.Post("/verify-receipt", iapRouter.VerifyReceipt)
		r.With(controller.FtcID).Post("/link", iapRouter.Link)
		r.With(controller.FtcID).Delete("/link", iapRouter.Unlink)

		r.Get("/receipt/{id}", iapRouter.LoadReceipt)
	})

	r.Route("/upgrade", func(r chi.Router) {
		r.Use(guard.CheckToken)
		r.Use(controller.UserOrUnionID)
		// Get membership information when user want to upgrade: days remaining, account balance, amount
		r.Put("/free", upgradeRouter.FreeUpgrade)
		r.Get("/balance", upgradeRouter.UpgradeBalance)
	})

	// Deprecate. Use /webhook
	r.Route("/callback", func(r1 chi.Router) {
		r1.Post("/wxpay", wxRouter.WebHook)
		r1.Post("/alipay", aliRouter.WebHook)
	})

	r.Route("/webhook", func(r chi.Router) {
		r.Post("/wxpay", wxRouter.WebHook)
		r.Post("/alipay", aliRouter.WebHook)
		r.Post("/stripe", stripeRouter.WebHook)
		r.Post("/apple", iapRouter.WebHook)
	})

	r.Route("/paywall", func(r chi.Router) {
		r.Use(guard.CheckToken)

		r.Get("/", paywallRouter.LoadPaywall)
		r.Get("/__refresh", paywallRouter.BustCache)

		// Get promotion schedule, pricing plans and banner content
		// Deprecated
		r.Get("/default", controller.DefaultPaywall)
		// Deprecated
		r.Get("/current", paywallRouter.GetPaywall)

		// Get default pricing plans
		// Deprecated
		r.Get("/pricing/default", controller.DefaultPricing)
		// Deprecated
		r.Get("/pricing/current", paywallRouter.GetPricing)

		// Deprecated
		r.Get("/promo", paywallRouter.GetPromo)
	})

	r.Route("/wx", func(r chi.Router) {

		r.Route("/oauth", func(r chi.Router) {

			r.With(controller.RequireAppID).Post("/login", wxAuth.Login)

			r.With(controller.RequireAppID).Put("/refresh", wxAuth.Refresh)

			r.Get("/callback", wxAuth.WebCallback)
		})
	})

	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("subscription-api started")

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func status(w http.ResponseWriter, _ *http.Request) {

	data := struct {
		Version string `json:"version"`
		Build   string `json:"build"`
		Commit  string `json:"commit"`
		Live    bool   `json:"live"`
	}{
		Version: version,
		Build:   build,
		Commit:  commit,
		Live:    cfg.Live(),
	}

	_ = view.Render(w, view.NewResponse().NoCache().SetBody(data))
}
