package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/controller"
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

	wxRouter := controller.NewWxRouter(db, isProd)
	aliRouter := controller.NewAliRouter(db, isProd)
	memberRouter := controller.NewMemberRouter(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if !isProd {
		r.Use(controller.LogRequest)
	}

	// r.Use(controller.CheckUserID)

	r.Use(controller.NoCache)

	r.Get("/__version", controller.Version(version, build))

	// Requires user id.
	r.Route("/wxpay", func(r1 chi.Router) {
		r1.Use(controller.CheckUserID)

		r1.Post("/unified-order/{tier}/{cycle}", wxRouter.UnifiedOrder)

		r1.Get("/order/{orderId}", wxRouter.OrderQuery)
	})

	// Requries user id.
	r.Route("/alipay", func(r1 chi.Router) {
		r1.Use(controller.CheckUserID)

		r1.Post("/app-order/{tier}/{cycle}", aliRouter.AppOrder)
		r1.Post("/verify/app-pay", aliRouter.VerifyAppPay)
	})

	r.Route("/membership", func(r1 chi.Router) {

		r1.Get("renewabel", memberRouter.IsRenewable)
	})

	r.Route("/callback", func(r1 chi.Router) {
		r1.Post("/wxpay", wxRouter.Notification)
		r1.Post("/alipay", aliRouter.Notification)
	})

	log.WithField("package", "subscription-api.main").Infof("subscription-api is running on port 8200")
	log.Fatal(http.ListenAndServe(":8200", r))
}
