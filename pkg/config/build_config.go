package config

import (
	"github.com/FTChinese/go-rest/connect"
	"github.com/spf13/viper"
	"log"
)

// BuildConfig set up deploy environment.
// When production is true, always connect to online db; otherwise connect to localhost.
// When sandbox is true, use stripe sandbox key; otherwise use test key.
// Wxpay and alipay's webhook url is also determined by sandbox. Their prices, however, is determined by
// current account.
// Local development: sandbox = true, production = false;
// Online price: sandbox = false, production = true;
// Online sandbox: sandbox = true, production = true;
// You should always run this program with `-sandbox` arg option.
type BuildConfig struct {
	sandboxMode bool // Determine the webhook base url for alipayw and wxpay. Determine stripe live/test mode.
	prodServer  bool // Determine which database should be used;
}

func NewBuildConfig(production, sandbox bool) BuildConfig {
	return BuildConfig{
		sandboxMode: sandbox,
		prodServer:  production,
	}
}

// Live determines which stripe key should be used.
// When the command line option is `-production`, it uses stripe live key;
// otherwise, it uses tripe test key: no options or `-production -sandbox`
func (c BuildConfig) Live() bool {
	return !c.sandboxMode
}

// Sandbox indicates API is running on production server in sandbox mode.
func (c BuildConfig) Sandbox() bool {
	return c.sandboxMode
}

// Production indicates API is running on production server.
func (c BuildConfig) Production() bool {
	return c.prodServer
}

// MustStripeSigningKey gets stripe signing key which is used to verify webhook data.
func (c BuildConfig) MustStripeSigningKey() string {
	var key string
	if c.sandboxMode {
		log.Print("Using stripe test signing key")
		key = viper.GetString("stripe.test_signing_key")
	} else {
		log.Print("Using stripe live signing key")
		key = viper.GetString("stripe.live_signing_key")
	}

	if key == "" {
		panic("cannot find stripe signing key")
	}

	return key
}

// MustStripeAPIKey gets stripe API key.
func (c BuildConfig) MustStripeAPIKey() string {
	var key string

	if c.sandboxMode {
		log.Print("Using stripe test api key")
		key = viper.GetString("stripe.test_secret_key")
	} else {
		log.Print("Using stripe live api key")
		key = viper.GetString("stripe.live_secret_key")
	}

	if key == "" {
		panic("cannot find stripe secret key")
	}

	return key
}

func MustGetHanqiConn() connect.Connect {
	conn, err := GetConn("email.hanqi")
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
