package main

import (
	"flag"
	"fmt"
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

	log.WithField("package", "next-api.main").Infof("Is production: %t", isProd)

	// NOTE: godotenv load .env file from current working directory, not where the program is located.
	err := godotenv.Load()
	if err != nil {
		log.WithField("package", "next-api.main").Error(err)
		os.Exit(1)
	}
}
func main() {
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASS")

	wx := util.WxConfig{}
	wx.AppID = os.Getenv("WXPAY_APPID")
	wx.MchID = os.Getenv("WXPAY_MCHID")
	wx.APIKey = os.Getenv("WXPAY_API_KEY")

	// log.WithField("package", "next-api.main").Infof("Connecting to MySQL: %s", host)

	db, err := util.NewDB(host, port, user, pass)
	if err != nil {
		// log.WithField("package", "next-api.main").Error(err)
		os.Exit(1)
	}

	orderRouter := controller.NewOrderRouter(wx, db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if !isProd {
		r.Use(controller.LogRequest)
	}

	r.Use(controller.CheckUserID)

	r.Use(controller.NoCache)

	r.Route("/place-order", func(r1 chi.Router) {
		r1.Post("/wxpay/{tier}/{cycle}", orderRouter.NewWxOrder)
	})

	// log.WithField("package", "subscription-api.main").Infof("subscription-api is running on port 8000")
	// log.Fatal(http.ListenAndServe(":8000", r))
}
