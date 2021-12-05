package apple

// The message corresponding to response status code.
// See https://developer.apple.com/documentation/appstorereceipts/status
var statusMessage = map[int64]string{
	21000: "The request to the App Store was not made using the HTTP POST request method",
	21001: "This status code is no longer sent by the App Store",
	21002: "The data in the receipt-data property was malformed or missing",
	21003: "The receipt could not be authenticated",
	21004: "The shared secret you provided does not match the shared secret on file for your account",
	21005: "The receipt server is not currently available",
	21006: "This receipt is valid but the subscription has expired",
	21007: "This receipt is from the test environment, but it was sent to the production environment for verification",
	21008: "This receipt is from the production environment, but it was sent to the test environment for verification",
	21009: "Internal data access error",
	21010: "The user account cannot be found or has been deleted",
}

func getStatusMessage(s int64) string {
	if s >= 21100 && s <= 21199 {
		return "Internal data access errors"
	}

	return statusMessage[s]
}
