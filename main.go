package main

import (
	"flag"
	"fmt"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/ali"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"gitlab.com/ftchinese/subscription-api/repository/iaprepo"
	"gitlab.com/ftchinese/subscription-api/repository/rederrepo"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"gitlab.com/ftchinese/subscription-api/repository/wxoauth"
	"log"
	"net/http"
	"os"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/controller"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

var (
	config  util.BuildConfig
	version string
	build   string
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

	config = util.NewBuildConfig(production, sandbox)

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.WithFields(logrus.Fields{
		"sandbox": config.UseSandboxDB(),
		"live":    config.Live(),
	}).Infof("Initializing environment")

	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(1)
	}

	if config.Live() {
		stripe.Key = viper.GetString("stripe.live_secret_key")
	} else {
		stripe.Key = viper.GetString("stripe.test_secret_key")
	}
}

func getStripeSecretKey() string {
	var key string

	if config.Live() {
		key = viper.GetString("stripe.live_secret_key")
	} else {
		key = viper.GetString("stripe.test_secret_key")
	}

	if key == "" {
		logrus.Error("cannot find stripe secret key")
		os.Exit(1)
	}

	return key
}

func getStripeSigningKey() string {
	var key string
	if config.Live() {
		key = viper.GetString("stripe.live_signing_key")
	} else {
		key = viper.GetString("stripe.test_signing_key")
	}

	if key == "" {
		logrus.Error("cannot find stripe signing key")
		os.Exit(1)
	}

	return key
}

func getDBConn() util.Conn {
	// Get DB connection config.
	var conn util.Conn
	var err error
	if config.Live() {
		err = viper.UnmarshalKey("mysql.master", &conn)
	} else {
		err = viper.UnmarshalKey("mysql.dev", &conn)
	}

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	logrus.Infof("Using MySQL server %s", conn.Host)
	return conn
}

func getEmailConn() util.Conn {
	var conn util.Conn
	err := viper.UnmarshalKey("email.hanqi", &conn)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return conn
}

func main() {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "main",
	})

	stripe.Key = getStripeSecretKey()

	db, err := util.NewDB(getDBConn())
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	c := cache.New(cache.DefaultExpiration, 0)

	emailConn := getEmailConn()
	post := postoffice.NewPostman(
		emailConn.Host,
		emailConn.Port,
		emailConn.User,
		emailConn.Pass)

	subEnv := subrepo.NewSubEnv(db, c, config)
	readerEnv := rederrepo.NewReaderEnv(db)
	iapEnv := iaprepo.NewIAPEnv(db, config)

	baseRouter := controller.NewBasePayRouter(subEnv, readerEnv, post)

	wxRouter := controller.NewWxRouter(baseRouter)
	aliRouter := controller.NewAliRouter(baseRouter)
	stripeRouter := controller.NewStripeRouter(baseRouter, getStripeSigningKey())
	iapRouter := controller.NewIAPRouter(iapEnv, readerEnv, post)

	giftCardRouter := controller.NewGiftCardRouter(subEnv)
	paywallRouter := controller.NewPaywallRouter(subEnv)
	upgradeRouter := controller.NewUpgradeRouter(subEnv)

	wxAuth := controller.NewWxAuth(wxoauth.New(db))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
	r.Use(controller.NoCache)

	r.Get("/__version", status)
	// Inspect what pricing plans are in effect.
	r.Get("/__refresh", paywallRouter.RefreshPromo)

	// Requires user id.
	r.Route("/wxpay", func(r chi.Router) {
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
		r.Use(controller.UserOrUnionID)

		r.Post("/desktop/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryDesktopWeb))

		r.Post("/mobile/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryMobileWeb))

		r.Post("/app/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryApp))

		// Deprecate
		//r.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		// r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
	})

	r.Route("/stripe", func(r chi.Router) {
		r.Use(controller.FtcID)

		// Get a stripe plan.
		r.Get("/plans/{id}", stripeRouter.GetPlan)

		r.Route("/customers", func(r chi.Router) {
			// Create a stripe customer if not exists yet, or
			// just return the customer id if already exists.
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
		r.Post("/verify-receipt", iapRouter.VerifyReceipt)
		r.Delete("/link", iapRouter.Unlink)
	})

	r.Route("/upgrade", func(r chi.Router) {
		r.Use(controller.UserOrUnionID)
		// Get membership information when user want to upgrade: days remaining, account balance, amount
		// Deprecate
		//r.Put("/", upgradeRouter.DirectUpgrade)
		// Deprecate
		//r.Get("/preview", upgradeRouter.PreviewUpgrade)

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

	r.Route("/gift-card", func(r chi.Router) {
		r.Use(controller.UserOrUnionID)

		r.Put("/redeem", giftCardRouter.Redeem)
	})

	r.Route("/paywall", func(r chi.Router) {
		// Get promotion schedule, pricing plans and banner content
		r.Get("/default", controller.DefaultPaywall)
		r.Get("/current", paywallRouter.GetPaywall)

		// Get default pricing plans
		r.Get("/pricing/default", controller.DefaultPricing)
		r.Get("/pricing/current", paywallRouter.GetPricing)

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
		Sandbox bool   `json:"sandbox"`
	}{
		Version: version,
		Build:   build,
		Sandbox: config.UseSandboxDB(),
	}

	_ = view.Render(w, view.NewResponse().NoCache().SetBody(data))
}
