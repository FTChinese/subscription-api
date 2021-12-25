package wxpay

import (
	"errors"
	"fmt"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat"
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

func IsValidPayload(bm gopay.BodyMap) error {
	err := badRequestErr(bm.GetString(keyReturnCode), bm.GetString(keyReturnMsg))
	if err != nil {
		return err
	}

	err = unprocessableErr(
		bm.GetString(keyResultCode),
		bm.GetString(keyErrCode),
		bm.GetString(keyErrCodeDes))
	if err != nil {
		return err
	}

	if bm.GetString(keyAppID) == "" {
		return errors.New("wxpay response missing appid")
	}

	if bm.GetString(keyMchID) == "" {
		return errors.New("wxpay response missing mch_id")
	}

	if bm.GetString(keyNonceStr) == "" {
		return errors.New("wxpay response missing nonce_str")
	}

	if bm.GetString(keySign) == "" {
		return errors.New("wxpay response missing sign")
	}

	return nil
}

func GetAppID(bm gopay.BodyMap) string {
	return bm.GetString(keyAppID)
}

func getSign(bm gopay.BodyMap) string {
	return bm.GetString(keySign)
}

func (app AppConfig) VerifySignature(bm gopay.BodyMap) (bool, error) {
	return wechat.VerifySign(
		app.APIKey,
		app.SignType,
		bm)
}

func (app AppConfig) ValidateOrderResponse(resp *wechat.UnifiedOrderResponse, o OrderParams) error {
	if err := badRequestErr(resp.ReturnCode, resp.ReturnMsg); err != nil {
		return err
	}

	if err := unprocessableErr(resp.ResultCode, resp.ErrCode, resp.ErrCodeDes); err != nil {
		return err
	}

	if o.AppID != resp.Appid || o.MchID != resp.MchId || o.Nonce != resp.NonceStr {
		return errors.New("wxpay response mismatching identity")
	}

	bm := Marshal(resp)

	ok, err := wechat.VerifySign(app.APIKey, app.SignType, bm)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New("signature verification failed")
	}

	return nil
}
