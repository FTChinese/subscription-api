package wepay

import (
	"strings"
	"testing"
)

func TestDecodeXML(t *testing.T) {
	resp := createNotificationResponse()

	p := DecodeXML(strings.NewReader(resp))

	t.Logf("Wechat notification: %+v\n", p)
}

func TestParseResponse(t *testing.T) {
	resp := createNotificationResponse()

	p, err := ParseResponse(mockClient, strings.NewReader(resp))

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Wechat notification: %+v\n", p)
}

func TestValidateResponse(t *testing.T) {
	resp := createNotificationResponse()

	p, err := ParseResponse(mockClient, strings.NewReader(resp))

	if err != nil {
		t.Error(err)
		return
	}

	reason := ValidateResponse(p)

	if reason != nil {
		t.Logf("Valiation failed: %+v\n", reason)
	}
}
