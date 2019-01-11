package controller

import (
	"strings"
	"testing"

	"gitlab.com/ftchinese/subscription-api/wepay"
)

func TestCreateNotification(t *testing.T) {
	n := createNotification()

	t.Logf("A mock notification: %s\n", n)
}

func TestProcessResponse(t *testing.T) {
	n := createNotification()
	resp := strings.NewReader(n)
	params, err := wepay.ParseResponse(mockClient, resp)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parsed response: %+v\n", params)
}
