package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/go-chi/chi"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"net/http"
)

const (
	wxOAuthCallback = "http://users.ftchinese.com/login/wechat/callback?"
	wxAppNativeApp  = "wxacddf1c20516eb69" // Used by native app to pay and log in.
)

// getURLParam gets a url parameter.
func getURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

func getEdition(req *http.Request) (product.Edition, error) {
	t, err := getURLParam(req, "tier").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	tier, err := enum.ParseTier(t)
	if err != nil {
		return product.Edition{}, err
	}

	c, err := getURLParam(req, "cycle").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	cycle, err := enum.ParseCycle(c)
	if err != nil {
		return product.Edition{}, err
	}

	return product.Edition{
		Tier:  tier,
		Cycle: cycle,
	}, nil
}

// gatherWxPayInput collect request input data. Due to legacy issues the data is scattered in various places:
// tier and cycle in url param;
// openId in request body for wechat-specific browser pay.
// planId is not provided yet.
// Since client is not required to send a json body, json.Unmarshal might have problems when request body is empty.
// To had better stick to the old way and create new endpoint to ask client to submit all data in json body..
func gatherWxPayInput(platform wechat.TradeType, req *http.Request) (subs.WxPayInput, error) {
	input := subs.NewWxPayInput(platform)

	// Get the OpenID field.
	openID, err := GetJSONString(req.Body, "openId")
	if err != nil {
		return input, err
	}

	// Get the tier and cycle field
	edition, err := getEdition(req)
	if err != nil {
		return input, err
	}

	input.OpenID = null.NewString(openID, openID != "")
	input.Edition = edition

	return input, nil
}

func gatherAliPayInput(req *http.Request) (subs.AliPayInput, error) {
	var input subs.AliPayInput

	retUrl := req.FormValue("return_url")

	edition, err := getEdition(req)
	if err != nil {
		return input, err
	}

	input.Edition = edition
	input.ReturnURL = retUrl

	return input, nil
}

// getReaderIDs extract ftc uuid or union id from request header.
// It panic if both ftc id and union id are missing.
// However it won't happen since middlewares already ensured at least one of them should exist.
func getReaderIDs(h http.Header) reader.MemberID {
	ftcID := h.Get(ftcIDKey)
	unionID := h.Get(unionIDKey)

	return reader.MemberID{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}

// castStripeError tries to cast an error to stripe.Error, or nil if it is not.
func castStripeError(err error) *stripe.Error {
	if stripeErr, ok := err.(*stripe.Error); ok {
		return stripeErr
	}

	return nil
}
