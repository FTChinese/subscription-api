package test

import (
	"github.com/smartwalle/alipay"
	"github.com/smartwalle/alipay/encoding"
	"testing"
)

func TestAliPayWebhook_URLValues(t *testing.T) {
	p := NewNPC()

	o := p.OrderBuilder().Build()

	payload := NewAlipayWebhook(o).URLValues()

	t.Logf("%s", payload)

	t.Logf("%s", payload.Encode())
}

func TestAliPayWebhook_SignedParams(t *testing.T) {

	params, err := NewAlipayWebhook(NewPersona().OrderBuilder().Build()).
		SignedParams()

	if err != nil {
		panic(err)
	}

	t.Logf("%v", params)

	t.Logf("%s", params.Encode())
}

func TestAliPayWebhook_Encode(t *testing.T) {
	s := NewAlipayWebhook(NewPersona().OrderBuilder().Build()).
		Encode()

	t.Logf("%s", s)
}

func TestVerifySignature(t *testing.T) {
	params, err := NewAlipayWebhook(NewPersona().OrderBuilder().Build()).
		SignedParams()

	if err != nil {
		panic(err)
	}

	ok, err := alipay.VerifySign(params, encoding.FormatPublicKey(AliApp.PublicKey))

	if err != nil {
		t.Error(err)
		return
	}

	if ok {
		t.Logf("Verification succedded")
	} else {
		t.Logf("Verification failed")
	}
}

func TestAlipayWebhook_MockPayload(t *testing.T) {

	p := NewNPC()

	o := p.OrderBuilder().Build()

	NewRepo().MustSaveOrder(o)

	payload := NewAlipayWebhook(o).Encode()

	t.Logf("%s", payload)
}
