package main

import (
	"flag"
	"fmt"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"net/http"
	"os"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/controller"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

var (
	isProd  bool
	sandbox bool
	version string
	build   string
	logger  = log.WithField("project", "subscription-api").WithField("package", "main")
)

func init() {
	flag.BoolVar(&isProd, "production", false, "Connect to production MySQL database if present. Default to localhost.")
	flag.BoolVar(&sandbox, "sandbox", false, "Use sandbox database to save subscription data if present.")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logger.Infof("Is sandbox: %t. Is production: %t", sandbox, isProd)

	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	// Get DB connection config.
	var dbConn util.Conn
	var err error
	if isProd {
		err = viper.UnmarshalKey("mysql.master", &dbConn)
	} else {
		err = viper.UnmarshalKey("mysql.dev", &dbConn)
	}

	if err != nil {
		logger.WithField("trace", "main").Error(err)
		os.Exit(1)
	}

	// Get email server config.
	var emailConn util.Conn
	err = viper.UnmarshalKey("email.hanqi", &emailConn)
	if err != nil {
		logger.WithField("trace", "main").Error(err)
		os.Exit(1)
	}

	db, err := util.NewDB(dbConn)
	if err != nil {
		logger.WithField("trace", "main").Error(err)
		os.Exit(1)
	}
	logger.
		WithField("trace", "main").
		Infof("Connected to MySQL server %s", dbConn.Host)

	c := cache.New(cache.DefaultExpiration, 0)
	post := postoffice.NewPostman(
		emailConn.Host,
		emailConn.Port,
		emailConn.User,
		emailConn.Pass)

	m := model.New(db, c, sandbox)

	wxRouter := controller.NewWxRouter(m, post, sandbox)
	aliRouter := controller.NewAliRouter(m, post, sandbox)
	paywallRouter := controller.NewPaywallRouter(m)

	wxAuth := controller.NewWxAuth(m)

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

		r.Post("/desktop/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeDesktop))

		r.Post("/mobile/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeMobile))

		// {code: string}
		r.Post("/jsapi/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeJSAPI))

		r.Post("/app/{tier}/{cycle}", wxRouter.PlaceOrder(wechat.TradeTypeApp))

		// Deprecate
		r.Post("/unified-order/{tier}/{cycle}", wxRouter.AppOrder)

		// Query order
		// X-App-Id
		r.Get("/query/{orderId}", wxRouter.OrderQuery)

		// Cancel order
	})

	// Require user id.
	r.Route("/alipay", func(r chi.Router) {
		r.Use(controller.UserOrUnionID)

		r.Post("/desktop/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryDesktopWeb))

		r.Post("/mobile/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryMobileWeb))

		r.Post("/app/{tier}/{cycle}", aliRouter.PlaceOrder(ali.EntryApp))

		// Deprecate
		r.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		// r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
	})

	r.Route("/callback", func(r1 chi.Router) {
		r1.Post("/wxpay", wxRouter.Notification)
		r1.Post("/alipay", aliRouter.Notification)
	})

	r.Route("/redirect", func(r chi.Router) {
		r.Get("/alipay/done", aliRouter.RedirectNextUser)
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

	logger.WithField("trace", "main").Infof("subscription-api is running on port 8200")
	log.Fatal(http.ListenAndServe(":8200", r))
}

func status(w http.ResponseWriter, req *http.Request) {

	data := struct {
		Version string `json:"version"`
		Build   string `json:"build"`
		Sandbox bool   `json:"sandbox"`
	}{
		Version: version,
		Build:   build,
		Sandbox: sandbox,
	}

	view.Render(w, view.NewResponse().NoCache().SetBody(data))
}
