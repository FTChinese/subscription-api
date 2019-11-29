package util

// RunEnv determines
type RunEnv int

const (
	RunEnvProduction RunEnv = 1 // Using production db
	RunEnvLocal             = 2 // Using localhost.
	RunEnvSandbox           = 4 // Using production db's sandbox tables
)

// BuildConfig set up deploy environment.
// For production server, the `-production` flag is passed from
// command line argument.
// Running environments:
// 1. On production server using production db;
// 2. On production server using production db but data are written to sandbox tables;
// 3. Local machine for development only.
type BuildConfig struct {
	Sandbox    bool // indicates the it is running on a production server so that production db is used while the application is used only for testing.
	Production bool // it determines which database should be used;
}

func (c BuildConfig) Live() bool {
	return c.Production && !c.Sandbox
}

func (c BuildConfig) GetReceiptVerificationURL() string {

	if c.Live() {

		return "https://buy.itunes.apple.com/verifyReceipt"
	}

	return "https://sandbox.itunes.apple.com/verifyReceipt"
}
