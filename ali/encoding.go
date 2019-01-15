package ali

import (
	"crypto"
	"encoding/base64"
	"strings"

	"github.com/smartwalle/alipay/encoding"
)

const (
	keySignNodeName = "sign"
	keyAppPayResp   = "alipay_trade_app_pay_response"
	signTypeRSA2    = "RSA2"
	signTypeRSA     = "RSA"
)

// This is used to parse Alibaba's stupid trick of using JSON as plain string and ordered hash.
// Copied from github.com/smartwalle/alipay/alipay.go
func extractAppPayResp(rawData string, nodeName string) string {
	var nodeIndex = strings.LastIndex(rawData, nodeName)

	var dataStartIndex = nodeIndex + len(nodeName) + 2
	var signIndex = strings.LastIndex(rawData, "\""+keySignNodeName+"\"")
	var dataEndIndex = signIndex - 1

	var indexLen = dataEndIndex - dataStartIndex
	if indexLen < 0 {
		return ""
	}
	return rawData[dataStartIndex:dataEndIndex]
}

func verifyAliResp(data []byte, sign, singType string, key []byte) (ok bool, err error) {
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}

	if singType == signTypeRSA {
		err = encoding.VerifyPKCS1v15(data, signBytes, key, crypto.SHA1)
	} else {
		err = encoding.VerifyPKCS1v15(data, signBytes, key, crypto.SHA256)
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
