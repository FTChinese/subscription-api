package util

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
	if c.sandbox {
		return true
	}

	if c.production {
		return false
	}

	// Not sandbox, not production, it should be local development. Use the same environment as production.
	return false
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
