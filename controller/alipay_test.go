package controller

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/mitchellh/go-homedir"

	"github.com/smartwalle/alipay/encoding"
)

const data = `{"alipay_trade_app_pay_response":{"code":"10000","msg":"Success","app_id":"2018053060263354","auth_app_id":"2018053060263354","charset":"utf-8","timestamp":"2018-10-28 16:49:31","out_trade_no":"FT0055301540716534","total_amount":"0.01","trade_no":"2018102822001439881007782559","seller_id":"2088521304936335"},"sign":"MHrLSKA3KUKxsN9Yuhnzqbj5jpnSQ8drar5nt3gQJ0OzSTmmaYvYhEPEf/Qf6T+3t4UAnWmbRRGuHqruDK2/AuH+xtmhElPFLXo9dnkduUe5c15/AKtW6V2SWs+TGmSi38Wb/3NgeINtlSSxGnLXsW3uzbnybEd0E/L4nyqaKZ+yF3GWsWAsLzgf/O9y5ntpc7st3Vu1I2icipp34N9a4UbnOML0/kPuLls09K6/w461AAXh2GE4+L103lp/M4QFd5Lghauod75VctKI/xro06jIEjRkojFOOry+dugqEDxUQX+3CHzqOojub6ozD5GTZUV0ynOZCQA4iX+oOZ52lw==","sign_type":"RSA2"}`

const filePath = "~/go/src/gitlab.com/ftchinese/subscription-api/alipay_public_key.pem"

func TestHomeDir(t *testing.T) {
	keyPath, err := homedir.Expand(filePath)

	if err != nil {
		t.Error(err)
	}

	t.Log(keyPath)
}

func TestVerifyAppPay(t *testing.T) {
	keyPath, err := homedir.Expand(filePath)

	if err != nil {
		t.Error(err)

		return
	}

	keyStr, err := ioutil.ReadFile(keyPath)

	publicKey := encoding.ParsePublicKey(string(keyStr))

	signedStr := extractAppPayResp(data, keyAppPayResp)

	var result aliAppPayResult

	if err := json.Unmarshal([]byte(data), &result); err != nil {
		t.Error(err)
		return
	}

	sign := result.Sign
	signType := result.SignType

	ok, err := verifyAliResp([]byte(signedStr), sign, signType, publicKey)

	if err != nil {
		t.Error(err)
	}

	t.Log(ok)
}

func TestExtractAliSignedStr(t *testing.T) {
	var rootNodeName = "alipay_trade_app_pay_response"

	content := extractAppPayResp(data, rootNodeName)

	t.Log(content)
}
