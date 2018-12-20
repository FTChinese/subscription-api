package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/patrickmn/go-cache"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/controller"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

var (
	isProd  bool
	version string
	build   string
)

func init() {
	flag.BoolVar(&isProd, "production", false, "Indicate productions environment if present")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.WithField("package", "subscription-api.main").Infof("Is production: %t", isProd)

	// NOTE: godotenv load .env file from current working directory, not where the program is located.
	err := godotenv.Load()
	if err != nil {
		log.WithField("package", "subscription-api.main").Error(err)
		os.Exit(1)
	}
}
func main() {
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASS")

	log.WithField("package", "subscription-api.main").Infof("Connecting to MySQL: %s", host)

	db, err := util.NewDB(host, port, user, pass)
	if err != nil {
		log.WithField("package", "subscription-api.main").Error(err)
		os.Exit(1)
	}

	c := cache.New(cache.DefaultExpiration, 0)
	p := util.NewPostOffice()
	m := model.Env{DB: db, Cache: c, PostOffice: p}

	wxRouter := controller.NewWxRouter(m, isProd)
	aliRouter := controller.NewAliRouter(m, isProd)
	paywallRouter := controller.NewPaywallRouter(m)

	wxAuth := controller.NewWxAuth(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if !isProd {
		r.Use(controller.LogRequest)
	}

	// r.Use(controller.CheckUserID)

	r.Use(controller.NoCache)

	r.Get("/__version", controller.Version(version, build))
	// Inspect what pricing plans are in effect.
	r.Get("/__refresh", paywallRouter.RefreshPromo)

	r.Get("/__current_plans", paywallRouter.CurrentPlans)

	// Requires user id.
	r.Route("/wxpay", func(r1 chi.Router) {
		r1.Use(controller.CheckUserID)

		r1.Post("/unified-order/{tier}/{cycle}", wxRouter.UnifiedOrder)

		// Query order
		r1.Get("/query/{orderId}", wxRouter.OrderQuery)

		// Cancel order
	})

	// Requries user id.
	r.Route("/alipay", func(r1 chi.Router) {
		r1.Use(controller.CheckUserID)

		r1.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
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

	r.Route("/oauth", func(r1 chi.Router) {
		r1.Post("/wx-access", wxAuth.Login)
	})

	r.Route("/wx", func(r1 chi.Router) {
		r1.Route("/oauth", func(r2 chi.Router) {
			r2.Post("/mobile", wxAuth.Login)
			r2.Get("/callback", wxAuth.WebCallback)
		})

		r1.Route("/user", func(r2 chi.Router) {
			// All routers under this section should contain X-Union-Id header.
			r2.Use(controller.CheckUnionID)

			r2.Put("/bind", wxAuth.BindAccount)
			// r2.Delete("/bind")
			r2.Get("/account", wxAuth.LoadAccount)
		})
	})

	log.WithField("package", "subscription-api.main").Infof("subscription-api is running on port 8200")
	log.Fatal(http.ListenAndServe(":8200", r))
}
