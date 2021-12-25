package wechat

import (
	"errors"
	"fmt"
	"github.com/objcoding/wxpay"
)

func badRequestErr(code string, msg string) error {
	if code != Success {
		return fmt.Errorf("wxpay api bad requeest: %s, %s", code, msg)
	}
	return nil
}

func unprocessableErr(resultCode, errCode, errCodeDes string) error {
	if resultCode != Success {
		return fmt.Errorf("wxpay api unprocessable: %s - %s", errCode, errCodeDes)
	}

	return nil
}

func IsValidPayload(p wxpay.Params) error {
	err := badRequestErr(p.GetString(keyReturnCode), p.GetString(keyReturnMsg))
	if err != nil {
		return err
	}

	err = unprocessableErr(
		p.GetString(keyResultCode),
		p.GetString(keyErrCode),
		p.GetString(keyErrCodeDes))
	if err != nil {
		return err
	}

	if p.GetString(keyAppID) == "" {
		return errors.New("wxpay response missing appid")
	}

	if p.GetString(keyMchID) == "" {
		return errors.New("wxpay response missing mch_id")
	}

	if p.GetString(keyNonceStr) == "" {
		return errors.New("wxpay response missing nonce_str")
	}

	if p.GetString(keySign) == "" {
		return errors.New("wxpay response missing sign")
	}

	return nil
}

func (app PayApp) ValidateOrderPayload(payload wxpay.Params) error {
	if err := IsValidPayload(payload); err != nil {
		return err
	}

	if GetAppID(payload) != app.AppID {
		return errors.New("wxpay api response: appid mismatched")
	}

	if getMchID(payload) != app.MchID {
		return errors.New("wxpay api response: mch_id mismatched")
	}

	return nil
}

func ValidateWebhookPayload(p wxpay.Params) error {
	if err := IsValidPayload(p); err != nil {
		return err
	}

	if p[keyTotalAmount] == "" {
		return errors.New("no payment amount found in wx webhook")
	}

	if p[keyOrderID] == "" {
		return errors.New("no order id in wx webhook")
	}

	return nil
}
