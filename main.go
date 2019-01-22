package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/FTChinese/go-rest/view"
	cache "github.com/patrickmn/go-cache"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/controller"
	"gitlab.com/ftchinese/subscription-api/util"
)

var (
	sandbox bool
	version string
	build   string
)

func init() {
	flag.BoolVar(&sandbox, "sandbox", false, "Indicate production environment if present")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.WithField("project", "subscription-api").Infof("Is sandbox: %t", sandbox)

	// NOTE: godotenv load .env file from current working directory, not where the program is located.
	err := godotenv.Load()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
func main() {
	logger := log.WithField("project", "subscription-api").WithField("package", "main")

	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASS")

	logger.Infof("Connecting to MySQL: %s", host)

	db, err := util.NewDB(host, port, user, pass)
	if err != nil {
		logger.WithField("trace", "main").Error(err)
		os.Exit(1)
	}

	c := cache.New(cache.DefaultExpiration, 0)

	wxRouter := controller.NewWxRouter(db, c, sandbox)
	aliRouter := controller.NewAliRouter(db, c, sandbox)
	paywallRouter := controller.NewPaywallRouter(db, c, sandbox)

	wxAuth := controller.NewWxAuth(db, c)

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
		r1.Use(controller.CheckUserID)

		r1.Post("/unified-order/{tier}/{cycle}", wxRouter.UnifiedOrder)

		// Query order
		// r1.Get("/query/{orderId}", wxRouter.OrderQuery)

		// Cancel order
	})

	// Requries user id.
	r.Route("/alipay", func(r1 chi.Router) {
		r1.Use(controller.CheckUserID)

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

	r.Route("/wx", func(r1 chi.Router) {
		r1.Route("/oauth", func(r2 chi.Router) {
			r2.Post("/login", wxAuth.Login)
			r2.Put("/refresh", wxAuth.Refresh)
			r2.Get("/callback", wxAuth.WebCallback)
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
