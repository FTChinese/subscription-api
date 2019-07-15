package model

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

// Create a customer to be used for postman testing.
func TestAPI_NewUser(t *testing.T) {
	user := test.NewProfile().FtcUser()

	t.Logf("Customer %+v", user)
}
