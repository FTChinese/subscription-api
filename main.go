package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/controller"
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

func wxOAuthApps() map[string]wxlogin.WxApp {
	var mSubs, mFTC, wFTC wxlogin.WxApp

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	err := viper.UnmarshalKey("wxapp.m_subs", &mSubs)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mSubs.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app Member subscription has empty fields")
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	err = viper.UnmarshalKey("wxapp.m_ftc", &mFTC)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mFTC.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app FTC has empty fields")
		os.Exit(1)
	}
	// 网站应用 -> FT中文网. This is used for web login
	err = viper.UnmarshalKey("wxapp.w_ftc", &wFTC)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if wFTC.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Web app FTC has empty fields")
		os.Exit(1)
	}

	return map[string]wxlogin.WxApp{
		// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
		"***REMOVED***": mSubs,
		// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
		"***REMOVED***": mFTC,
		// 网站应用 -> FT中文网. This is used for web login
		"wxc7233549ca6bc86a": wFTC,
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
		logger.WithField("trace", "main").Error((err))
		os.Exit(1)
	}

	// Get email server config.
	var emailConn util.Conn
	err = viper.UnmarshalKey("hanqi", &emailConn)
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

	wxAuth := controller.NewWxAuth(m, wxOAuthApps())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(controller.LogRequest)
	r.Use(controller.NoCache)

	r.Get("/__version", status)
	// Inspect what pricing plans are in effect.
	r.Get("/__refresh", paywallRouter.RefreshPromo)

	r.Get("/__current_plans", paywallRouter.CurrentPlans)

	// Requires user id.
	r.Route("/wxpay", func(r1 chi.Router) {
		r1.Use(controller.UserOrUnionID)

		r1.Post("/unified-order/{tier}/{cycle}", wxRouter.UnifiedOrder)

		// Query order
		// r1.Get("/query/{orderId}", wxRouter.OrderQuery)

		// Cancel order
	})

	// Require user id.
	r.Route("/alipay", func(r1 chi.Router) {
		r1.Use(controller.UserOrUnionID)

		r1.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		// r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
	})

	r.Route("/callback", func(r1 chi.Router) {
		r1.Post("/wxpay", wxRouter.Notification)
		r1.Post("/alipay", aliRouter.Notification)
	})

	r.Route("/paywall", func(r1 chi.Router) {
		// Get promotion schedule, pricing plans and banner content
		r1.Get("/promo", paywallRouter.GetPromo)

		// Get products list
		// r1.Get("/products", )
		// Get default pricing plans
		r1.Get("/plans", controller.DefaultPlans)

		// Get default banner
		// r1.Get("/banner", )
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
