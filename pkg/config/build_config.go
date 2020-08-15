package config

import (
	"github.com/FTChinese/go-rest/connect"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
)

func GetConn(key string) (connect.Connect, error) {
	var conn connect.Connect
	err := viper.UnmarshalKey(key, &conn)
	if err != nil {
		return connect.Connect{}, err
	}

	return conn, nil
}

// BuildConfig set up deploy environment.
// For production server, the `-production` flag is passed from
// command line argument.
// Running environments:
// 1. On production server using production db;
// 2. On production server using production db but data are written to sandbox tables;
// 3. Local machine for development only.
type BuildConfig struct {
	sandbox    bool // indicates the it is running on a production server so that production db is used while the application is used only for testing.
	production bool // it determines which database should be used;
}

func NewBuildConfig(production, sandbox bool) BuildConfig {
	return BuildConfig{
		sandbox:    sandbox,
		production: production,
	}
}

// Live specifies:
// * Which stripe key should be used: true for live key;
// * How much should user pay: true for normal, false for 1 cent;
// * Which webhook url should be used: true for produciton url.
// Matrix to determine what to use: (test == sandbox)
// 				Stripe Key 	WebHook		DB		Price
// Local		test		sandbox		live	sandbox
// Production	live		live		live	live
// Sandbox		test		sandbox		sandbox	sandbox
func (c BuildConfig) Live() bool {
	return c.production && !c.sandbox
}

// UseSandboxDB tells whether the sandbox db should be used.
// Not this is not the opposite of Live.
func (c BuildConfig) UseSandboxDB() bool {
	return c.sandbox
}

// IsProduction determines which DB server to connect
func (c BuildConfig) IsProduction() bool {
	return c.production
}

// GetReceiptVerificationURL selects apple receipt verification
// endpoint depending on the deployment environment.
// This is the same to stripe key selection.
// MUST not use the UsedSandboxDB!
func (c BuildConfig) GetReceiptVerificationURL() string {

	if c.Live() {
		return "https://buy.itunes.apple.com/verifyReceipt"
	}

	return "https://sandbox.itunes.apple.com/verifyReceipt"
}

func (c BuildConfig) GetStripeKey() string {
	var key string
	if c.Live() {
		key = viper.GetString("stripe.live_signing_key")
	} else {
		key = viper.GetString("stripe.test_signing_key")
	}

	if key == "" {
		logrus.WithField("trace", "BuildConfig.GetStripeKey").
			Error("cannot find stripe signing key")
		os.Exit(1)
	}

	return key
}

func (c BuildConfig) GetStripeSecretKey() string {
	var key string

	if c.Live() {
		key = viper.GetString("stripe.live_secret_key")
	} else {
		key = viper.GetString("stripe.test_secret_key")
	}

	if key == "" {
		logrus.WithField("trace", "BuildConfig.GetStripeSecretKey").
			Error("cannot find stripe secret key")

		os.Exit(1)
	}

	return key
}

func (c BuildConfig) MustGetDBConn(key string) connect.Connect {
	var conn connect.Connect
	var err error

	if c.production {
		conn, err = GetConn(key)
	} else {
		conn, err = GetConn("mysql.dev")
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using mysql server %s. Production: %t", conn.Host, c.production)

	return conn
}

func MustGetHanqiConn() connect.Connect {
	conn, err := GetConn("email.hanqi")
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
